package internal

import (
	"context"
	"errors"
	"fmt"
	"github.com/googleapis/gax-go/v2"

	aiplatform "cloud.google.com/go/aiplatform/apiv1"
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/file"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/util"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ScheduleClient interface {
	CreateSchedule(
		ctx context.Context,
		req *aiplatformpb.CreateScheduleRequest,
		opts ...gax.CallOption,
	) (*aiplatformpb.Schedule, error)
	DeleteSchedule(
		ctx context.Context,
		req *aiplatformpb.DeleteScheduleRequest,
		opts ...gax.CallOption,
	) (*aiplatform.DeleteScheduleOperation, error)
}

type VAIProvider struct {
	ctx            context.Context
	config         VAIProviderConfig
	fileHandler    file.FileHandler
	pipelineClient PipelineJobClient
	scheduleClient ScheduleClient
	jobBuilder     JobBuilder
	jobEnricher    JobEnricher
}

func NewProvider(
	ctx context.Context,
	config VAIProviderConfig,
) (*VAIProvider, error) {
	fh, err := file.NewGcsFileHandler(ctx, config.Parameters.GcsEndpoint)
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
			serviceAccount: config.Parameters.VaiJobServiceAccount,
			pipelineBucket: config.Parameters.PipelineBucket,
			labelGen:       DefaultLabelGen{},
		},
		jobEnricher: DefaultJobEnricher{},
	}, nil
}

func (vaip *VAIProvider) CreatePipeline(
	pd resource.PipelineDefinition,
) (string, error) {
	pipelineId, err := vaip.UpdatePipeline(pd, "")
	if err != nil {
		return "", err
	}
	return pipelineId, nil
}

func (vaip *VAIProvider) UpdatePipeline(
	pd resource.PipelineDefinition,
	_ string,
) (string, error) {
	pipelineId, err := pd.Name.String()
	if err != nil {
		return "", err
	}

	storageObject, err := util.PipelineStorageObject(pd.Name, pd.Version)
	if err != nil {
		return pipelineId, err
	}

	if err = vaip.fileHandler.Write(
		pd.Manifest,
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

	req := &aiplatformpb.CreatePipelineJobRequest{
		Parent:        vaip.config.parent(),
		PipelineJobId: fmt.Sprintf("%s-%s", runId, rd.Version),
		PipelineJob:   enrichedJob,
	}

	_, err = vaip.pipelineClient.CreatePipelineJob(vaip.ctx, req)
	if err != nil {
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
		vaip.config.parent(),
		vaip.config.getMaxConcurrentRunCountOrDefault(),
	)
	if err != nil {
		return "", err
	}

	createdSchedule, err := vaip.scheduleClient.CreateSchedule(
		vaip.ctx,
		&aiplatformpb.CreateScheduleRequest{
			Parent:   vaip.config.parent(),
			Schedule: schedule,
		},
	)
	if err != nil {
		return "", err
	}

	return createdSchedule.Name, nil
}

func (vaip *VAIProvider) UpdateRunSchedule(
	rsd resource.RunScheduleDefinition,
	id string,
) (string, error) {
	return "", nil
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
