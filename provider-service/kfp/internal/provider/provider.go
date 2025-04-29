package provider

import (
	"context"
	baseResource "github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type KfpProvider struct {
	config                *config.KfpProviderConfig
	pipelineUploadService PipelineUploadService
	pipelineService       PipelineService
	runService            RunService
	experimentService     ExperimentService
	jobService            JobService
}

func NewKfpProvider(
	providerConfig *config.KfpProviderConfig,
) (*KfpProvider, error) {
	pipelineUploadService, err := NewPipelineUploadService(
		providerConfig.Parameters.RestKfpApiUrl,
	)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(
		providerConfig.Parameters.GrpcKfpApiAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	pipelineService, err := NewPipelineService(conn)
	if err != nil {
		return nil, err
	}

	runService, err := NewRunService(conn)
	if err != nil {
		return nil, err
	}

	experimentService, err := NewExperimentService(conn)
	if err != nil {
		return nil, err
	}

	jobService, err := NewJobService(conn)
	if err != nil {
		return nil, err
	}

	return &KfpProvider{
		config:                providerConfig,
		pipelineUploadService: pipelineUploadService,
		pipelineService:       pipelineService,
		runService:            runService,
		experimentService:     experimentService,
		jobService:            jobService,
	}, nil
}

func (kfpp *KfpProvider) CreatePipeline(
	ctx context.Context,
	pdw baseResource.PipelineDefinitionWrapper,
) (string, error) {
	pipelineName, err := pdw.PipelineDefinition.Name.String()
	if err != nil {
		return "", err
	}

	pipelineId, err := kfpp.pipelineUploadService.UploadPipeline(
		ctx,
		pdw.CompiledPipeline,
		pipelineName,
	)
	if err != nil {
		return "", err
	}

	return kfpp.UpdatePipeline(ctx, pdw, pipelineId)
}

func (kfpp *KfpProvider) UpdatePipeline(
	ctx context.Context,
	pdw baseResource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	if err := kfpp.pipelineUploadService.UploadPipelineVersion(
		ctx,
		id,
		pdw.CompiledPipeline,
		pdw.PipelineDefinition.Version,
	); err != nil {
		return "", err
	}

	return id, nil
}

func (kfpp *KfpProvider) DeletePipeline(ctx context.Context, id string) error {
	return kfpp.pipelineService.DeletePipeline(ctx, id)
}

func (kfpp *KfpProvider) CreateRun(ctx context.Context, rd baseResource.RunDefinition) (string, error) {
	pipelineName, err := util.ResourceNameFromNamespacedName(rd.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := kfpp.pipelineService.PipelineIdForName(ctx, pipelineName)
	if err != nil {
		return "", err
	}

	pipelineVersionId, err := kfpp.pipelineService.PipelineVersionIdForName(
		ctx,
		rd.PipelineVersion,
		pipelineId,
	)
	if err != nil {
		return "", err
	}

	experimentId, err := kfpp.experimentService.ExperimentIdByName(ctx, rd.ExperimentName)
	if err != nil {
		return "", err
	}

	runId, err := kfpp.runService.CreateRun(
		ctx,
		rd,
		pipelineId,
		pipelineVersionId,
		experimentId,
	)
	if err != nil {
		return "", err
	}

	return runId, nil
}

func (kfpp *KfpProvider) DeleteRun(_ context.Context, _ string) error {
	// Not implemented for KFP provider
	// Required to satisfy the `Provider` interface
	return nil
}

func (kfpp *KfpProvider) CreateRunSchedule(
	ctx context.Context,
	rsd baseResource.RunScheduleDefinition,
) (string, error) {
	pipelineName, err := util.ResourceNameFromNamespacedName(rsd.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := kfpp.pipelineService.PipelineIdForName(ctx, pipelineName)
	if err != nil {
		return "", err
	}

	pipelineVersionId, err := kfpp.pipelineService.PipelineVersionIdForName(
		ctx,
		rsd.PipelineVersion,
		pipelineId,
	)
	if err != nil {
		return "", err
	}

	experimentId, err := kfpp.experimentService.ExperimentIdByName(ctx, rsd.ExperimentName)
	if err != nil {
		return "", err
	}

	jobId, err := kfpp.jobService.CreateJob(
		ctx,
		rsd,
		pipelineId,
		pipelineVersionId,
		experimentId,
	)
	if err != nil {
		return "", err
	}

	return jobId, nil
}

func (kfpp *KfpProvider) UpdateRunSchedule(
	ctx context.Context,
	rsd baseResource.RunScheduleDefinition,
	id string,
) (string, error) {
	if err := kfpp.DeleteRunSchedule(ctx, id); err != nil {
		return id, err
	}

	return kfpp.CreateRunSchedule(ctx, rsd)
}

func (kfpp *KfpProvider) DeleteRunSchedule(ctx context.Context, id string) error {
	return kfpp.jobService.DeleteJob(ctx, id)
}

func (kfpp *KfpProvider) CreateExperiment(
	ctx context.Context,
	ed baseResource.ExperimentDefinition,
) (string, error) {
	expId, err := kfpp.experimentService.CreateExperiment(
		ctx,
		ed.Name,
		ed.Description,
	)
	if err != nil {
		return "", err
	}

	return expId, nil
}

func (kfpp *KfpProvider) UpdateExperiment(
	ctx context.Context,
	ed baseResource.ExperimentDefinition,
	id string,
) (string, error) {
	if err := kfpp.DeleteExperiment(ctx, id); err != nil {
		return id, err
	}

	return kfpp.CreateExperiment(ctx, ed)
}

func (kfpp *KfpProvider) DeleteExperiment(ctx context.Context, id string) error {
	return kfpp.experimentService.DeleteExperiment(ctx, id)
}
