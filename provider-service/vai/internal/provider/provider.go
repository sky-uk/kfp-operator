package provider

import (
	"context"
	"errors"
	"fmt"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/util"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type VAIProvider struct {
	ctx            context.Context
	config         *config.VAIProviderConfig
	fileHandler    FileHandler
	pipelineClient client.PipelineJobClient
	scheduleClient client.ScheduleClient
	jobBuilder     JobBuilder
	jobEnricher    JobEnricher
}

func NewVAIProvider(
	ctx context.Context,
	config *config.VAIProviderConfig,
) (*VAIProvider, error) {
	fh, err := NewGcsFileHandler(ctx, config.Parameters.GcsEndpoint)
	if err != nil {
		return nil, err
	}

	pc, err := aiplatform.NewPipelineClient(
		ctx,
		option.WithEndpoint(config.VaiEndpoint()),
	)
	if err != nil {
		return nil, err
	}

	sc, err := aiplatform.NewScheduleClient(
		ctx,
		option.WithEndpoint(config.VaiEndpoint()),
	)
	if err != nil {
		return nil, err
	}

	return &VAIProvider{
		ctx:            ctx,
		config:         config,
		fileHandler:    &fh,
		pipelineClient: pc,
		scheduleClient: sc,
		jobBuilder: DefaultJobBuilder{
			serviceAccount:      config.Parameters.VaiJobServiceAccount,
			pipelineRootStorage: config.PipelineRootStorage,
			pipelineBucket:      config.Parameters.PipelineBucket,
			labelGen:            DefaultLabelGen{},
		},
		jobEnricher: DefaultJobEnricher{pipelineSchemaHandler: DefaultPipelineSchemaHandler{
			schema2Handler:   Schema2Handler{},
			schema2_1Handler: Schema2_1Handler{},
		}},
	}, nil
}

func (vaip *VAIProvider) CreatePipeline(
	pdw resource.PipelineDefinitionWrapper,
) (string, error) {
	pipelineId, err := vaip.UpdatePipeline(pdw, "")
	if err != nil {
		return "", err
	}
	return pipelineId, nil
}

func (vaip *VAIProvider) UpdatePipeline(
	pdw resource.PipelineDefinitionWrapper,
	_ string,
) (string, error) {
	pipelineId, err := pdw.PipelineDefinition.Name.String()
	if err != nil {
		return "", err
	}

	storageObject, err := util.PipelineStorageObject(
		pdw.PipelineDefinition.Name,
		pdw.PipelineDefinition.Version,
	)
	if err != nil {
		return pipelineId, err
	}

	if err = vaip.fileHandler.Write(
		pdw.CompiledPipeline,
		vaip.config.Parameters.PipelineBucket,
		storageObject,
	); err != nil {
		return "", err
	}
	return pipelineId, nil
}

func (vaip *VAIProvider) DeletePipeline(id string) error {
	if err := vaip.fileHandler.Delete(
		id,
		vaip.config.Parameters.PipelineBucket,
	); err != nil {
		return err
	}
	return nil
}

func (vaip *VAIProvider) CreateRun(rd resource.RunDefinition) (string, error) {
	logger := common.LoggerFromContext(vaip.ctx)

	pipelinePath, err := util.PipelineStorageObject(
		rd.PipelineName,
		rd.PipelineVersion,
	)
	if err != nil {
		return "", err
	}

	raw, err := vaip.fileHandler.Read(
		vaip.config.Parameters.PipelineBucket,
		pipelinePath,
	)
	if err != nil {
		return "", err
	}

	job, err := vaip.jobBuilder.MkRunPipelineJob(rd)
	if err != nil {
		return "", err
	}

	enrichedJob, err := vaip.jobEnricher.Enrich(job, raw)
	if err != nil {
		return "", err
	}

	runId := fmt.Sprintf("%s-%s", rd.Name.Namespace, rd.Name.Name)
	pipelineJobId := fmt.Sprintf("%s-%s", runId, rd.Version)

	req := &aiplatformpb.CreatePipelineJobRequest{
		Parent:        vaip.config.Parent(),
		PipelineJobId: pipelineJobId,
		PipelineJob:   enrichedJob,
	}

	_, err = vaip.pipelineClient.CreatePipelineJob(vaip.ctx, req)
	if err != nil {
		logger.Error(err, "CreatePipelineJob failed", "pipelineJobId", pipelineJobId)
		return "", err
	}

	return runId, nil
}

func (vaip *VAIProvider) DeleteRun(_ string) error {
	return nil
}

func (vaip *VAIProvider) CreateRunSchedule(
	rsd resource.RunScheduleDefinition,
) (string, error) {
	logger := common.LoggerFromContext(vaip.ctx)

	pipelinePath, err := util.PipelineStorageObject(
		rsd.PipelineName,
		rsd.PipelineVersion,
	)
	if err != nil {
		return "", err
	}

	raw, err := vaip.fileHandler.Read(
		vaip.config.Parameters.PipelineBucket,
		pipelinePath,
	)
	if err != nil {
		return "", err
	}

	job, err := vaip.jobBuilder.MkRunSchedulePipelineJob(rsd)
	if err != nil {
		return "", err
	}

	enrichedJob, err := vaip.jobEnricher.Enrich(job, raw)
	if err != nil {
		return "", err
	}

	schedule, err := vaip.jobBuilder.MkSchedule(
		rsd,
		enrichedJob,
		vaip.config.Parent(),
		vaip.config.GetMaxConcurrentRunCountOrDefault(),
	)
	if err != nil {
		return "", err
	}

	createdSchedule, err := vaip.scheduleClient.CreateSchedule(
		vaip.ctx,
		&aiplatformpb.CreateScheduleRequest{
			Parent:   vaip.config.Parent(),
			Schedule: schedule,
		},
	)
	if err != nil {
		logger.Error(err, "CreateScheduleRequest failed")
		return "", err
	}
	logger.Info("CreateScheduleRequest succeeded", "schedule name", createdSchedule.Name)

	return createdSchedule.Name, nil
}

func (vaip *VAIProvider) UpdateRunSchedule(
	rsd resource.RunScheduleDefinition,
	_ string,
) (string, error) {
	logger := common.LoggerFromContext(vaip.ctx)

	pipelinePath, err := util.PipelineStorageObject(
		rsd.PipelineName,
		rsd.PipelineVersion,
	)
	if err != nil {
		return "", err
	}

	raw, err := vaip.fileHandler.Read(
		vaip.config.Parameters.PipelineBucket,
		pipelinePath,
	)
	if err != nil {
		return "", err
	}

	job, err := vaip.jobBuilder.MkRunSchedulePipelineJob(rsd)
	if err != nil {
		return "", err
	}

	enrichedJob, err := vaip.jobEnricher.Enrich(job, raw)
	if err != nil {
		return "", err
	}

	schedule, err := vaip.jobBuilder.MkSchedule(
		rsd,
		enrichedJob,
		vaip.config.Parent(),
		vaip.config.GetMaxConcurrentRunCountOrDefault(),
	)
	if err != nil {
		return "", err
	}

	updateSchedule, err := vaip.scheduleClient.UpdateSchedule(
		vaip.ctx,
		&aiplatformpb.UpdateScheduleRequest{
			Schedule: schedule,
			UpdateMask: &fieldmaskpb.FieldMask{
				Paths: []string{
					"schedule",
				},
			},
		},
	)
	if err != nil {
		logger.Error(err, "UpdateScheduleRequest failed", "parent", vaip.config.Parent())
		return "", err
	}

	return updateSchedule.Name, nil
}

func (vaip *VAIProvider) DeleteRunSchedule(id string) error {
	schedule, err := vaip.scheduleClient.DeleteSchedule(
		vaip.ctx,
		&aiplatformpb.DeleteScheduleRequest{
			Name: id,
		},
	)
	if err != nil {
		return ignoreNotFound(err)
	}
	return ignoreNotFound(schedule.Wait(vaip.ctx))
}

func ignoreNotFound(err error) error {
	if status.Code(err) == codes.NotFound {
		return nil
	}
	return err
}

func (vaip *VAIProvider) CreateExperiment(
	_ resource.ExperimentDefinition,
) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip *VAIProvider) UpdateExperiment(
	_ resource.ExperimentDefinition,
	_ string,
) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip *VAIProvider) DeleteExperiment(_ string) error {
	return errors.New("not implemented")
}
