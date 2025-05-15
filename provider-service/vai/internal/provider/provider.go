package provider

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/client"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/config"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

type VAIProvider struct {
	config         *config.VAIProviderConfig
	fileHandler    FileHandler
	pipelineClient client.PipelineJobClient
	scheduleClient client.ScheduleClient
	jobBuilder     JobBuilder
	jobEnricher    JobEnricher
}

func (vaip *VAIProvider) Name() string {
	return "VaiProvider"
}

func (vaip VAIProvider) IsHealthy(ctx context.Context) bool {
	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		_, err := vaip.pipelineClient.ListPipelineJobs(ctx, &aiplatformpb.ListPipelineJobsRequest{
			Parent:   vaip.config.Parent(),
			PageSize: 1,
		}).Next()
		return err
	})

	g.Go(func() error {
		_, err := vaip.scheduleClient.ListSchedules(ctx, &aiplatformpb.ListSchedulesRequest{
			Parent:   vaip.config.Parent(),
			PageSize: 1,
		}).Next()
		return err
	})

	if err := g.Wait(); err != nil {
		return false
	}

	return true

}

func NewVAIProvider(
	ctx context.Context,
	config *config.VAIProviderConfig,
	namespace string,
	pipelineClient client.PipelineJobClient,
	scheduleClient client.ScheduleClient,
) (*VAIProvider, error) {
	fh, err := NewGcsFileHandler(ctx, config.Parameters.GcsEndpoint)
	if err != nil {
		return nil, err
	}

	return &VAIProvider{
		config:         config,
		fileHandler:    &fh,
		pipelineClient: pipelineClient,
		scheduleClient: scheduleClient,
		jobBuilder: DefaultJobBuilder{
			serviceAccount:      config.Parameters.VaiJobServiceAccount,
			pipelineRootStorage: config.PipelineRootStorage,
			pipelineBucket:      config.Parameters.PipelineBucket,
			labelGen:            DefaultLabelGen{providerName: common.NamespacedName{Name: config.Name, Namespace: namespace}},
		},
		jobEnricher: DefaultJobEnricher{pipelineSchemaHandler: DefaultPipelineSchemaHandler{
			schema2Handler:   Schema2Handler{},
			schema2_1Handler: Schema2_1Handler{},
		}},
	}, nil
}

func (vaip *VAIProvider) CreatePipeline(
	ctx context.Context,
	pdw resource.PipelineDefinitionWrapper,
) (string, error) {
	pipelineId, err := vaip.UpdatePipeline(ctx, pdw, "")
	if err != nil {
		return "", err
	}
	return pipelineId, nil
}

func (vaip *VAIProvider) UpdatePipeline(
	ctx context.Context,
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
		ctx,
		pdw.CompiledPipeline,
		vaip.config.Parameters.PipelineBucket,
		storageObject,
	); err != nil {
		return "", err
	}
	return pipelineId, nil
}

func (vaip *VAIProvider) DeletePipeline(ctx context.Context, id string) error {
	if err := vaip.fileHandler.Delete(
		ctx,
		id,
		vaip.config.Parameters.PipelineBucket,
	); err != nil {
		return err
	}
	return nil
}

func (vaip *VAIProvider) CreateRun(ctx context.Context, rd resource.RunDefinition) (string, error) {
	logger := common.LoggerFromContext(ctx)

	pipelinePath, err := util.PipelineStorageObject(
		rd.PipelineName,
		rd.PipelineVersion,
	)
	if err != nil {
		return "", err
	}

	raw, err := vaip.fileHandler.Read(
		ctx,
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

	_, err = vaip.pipelineClient.CreatePipelineJob(ctx, req)
	if err != nil {
		logger.Error(err, "CreatePipelineJob failed", "pipelineJobId", pipelineJobId)
		return "", err
	}

	return runId, nil
}

func (vaip *VAIProvider) DeleteRun(_ context.Context, _ string) error {
	return nil
}

func (vaip *VAIProvider) CreateRunSchedule(
	ctx context.Context,
	rsd resource.RunScheduleDefinition,
) (string, error) {
	logger := common.LoggerFromContext(ctx)

	pipelinePath, err := util.PipelineStorageObject(
		rsd.PipelineName,
		rsd.PipelineVersion,
	)
	if err != nil {
		return "", err
	}

	raw, err := vaip.fileHandler.Read(
		ctx,
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
		ctx,
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
	ctx context.Context,
	rsd resource.RunScheduleDefinition,
	_ string,
) (string, error) {
	logger := common.LoggerFromContext(ctx)

	pipelinePath, err := util.PipelineStorageObject(
		rsd.PipelineName,
		rsd.PipelineVersion,
	)
	if err != nil {
		return "", err
	}

	raw, err := vaip.fileHandler.Read(
		ctx,
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
		ctx,
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

func (vaip *VAIProvider) DeleteRunSchedule(ctx context.Context, id string) error {
	schedule, err := vaip.scheduleClient.DeleteSchedule(
		ctx,
		&aiplatformpb.DeleteScheduleRequest{
			Name: id,
		},
	)
	if err != nil {
		return ignoreNotFound(err)
	}
	return ignoreNotFound(schedule.Wait(ctx))
}

func ignoreNotFound(err error) error {
	if status.Code(err) == codes.NotFound {
		return nil
	}
	return err
}

func (vaip *VAIProvider) CreateExperiment(
	_ context.Context,
	_ resource.ExperimentDefinition,
) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip *VAIProvider) UpdateExperiment(
	_ context.Context,
	_ resource.ExperimentDefinition,
	_ string,
) (string, error) {
	return "", errors.New("not implemented")
}

func (vaip *VAIProvider) DeleteExperiment(
	_ context.Context,
	_ string,
) error {
	return errors.New("not implemented")
}
