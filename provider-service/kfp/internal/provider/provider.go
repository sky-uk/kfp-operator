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
	ctx                   context.Context
	config                *config.KfpProviderConfig
	pipelineUploadService PipelineUploadService
	pipelineService       PipelineService
	runService            RunService
	experimentService     ExperimentService
	jobService            JobService
}

func NewKfpProvider(
	ctx context.Context,
	providerConfig *config.KfpProviderConfig,
) (*KfpProvider, error) {
	pipelineUploadService, err := NewPipelineUploadService(
		ctx,
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

	pipelineService, err := NewPipelineService(ctx, conn)
	if err != nil {
		return nil, err
	}

	runService, err := NewRunService(ctx, conn)
	if err != nil {
		return nil, err
	}

	experimentService, err := NewExperimentService(ctx, conn)
	if err != nil {
		return nil, err
	}

	jobService, err := NewJobService(ctx, conn)
	if err != nil {
		return nil, err
	}

	return &KfpProvider{
		ctx:                   ctx,
		config:                providerConfig,
		pipelineUploadService: pipelineUploadService,
		pipelineService:       pipelineService,
		runService:            runService,
		experimentService:     experimentService,
		jobService:            jobService,
	}, nil
}

func (kfpp *KfpProvider) CreatePipeline(
	pdw baseResource.PipelineDefinitionWrapper,
) (string, error) {
	pipelineId, err := pdw.PipelineDefinition.Name.String()
	if err != nil {
		return "", err
	}

	//TODO: What should filePath be here???
	// it's not set anywhere so maybe it should just empty string
	result, err := kfpp.pipelineUploadService.UploadPipeline(
		pdw.CompiledPipeline,
		pipelineId,
		"resource.json",
	)
	if err != nil {
		return "", err
	}

	return kfpp.UpdatePipeline(pdw, result)
}

func (kfpp *KfpProvider) UpdatePipeline(
	pdw baseResource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	//TODO: What should filePath be here???
	// returning a result is pointless because it's just id again. Remove?
	if err := kfpp.pipelineUploadService.UploadPipelineVersion(
		id,
		pdw.CompiledPipeline,
		pdw.PipelineDefinition.Version,
		"resource.json",
	); err != nil {
		return "", err
	}

	return id, nil
}

func (kfpp *KfpProvider) DeletePipeline(id string) error {
	return kfpp.pipelineService.DeletePipeline(id)
}

func (kfpp *KfpProvider) CreateRun(rd baseResource.RunDefinition) (string, error) {
	pipelineName, err := util.ResourceNameFromNamespacedName(rd.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := kfpp.pipelineService.PipelineIdForName(pipelineName)
	if err != nil {
		return "", err
	}

	pipelineVersionId, err := kfpp.pipelineService.PipelineVersionIdForName(
		rd.PipelineVersion,
		pipelineId,
	)
	if err != nil {
		return "", err
	}

	experimentVersion, err := kfpp.experimentService.ExperimentIdByName(rd.ExperimentName)
	if err != nil {
		return "", err
	}

	runId, err := kfpp.runService.CreateRun(
		rd,
		pipelineId,
		pipelineVersionId,
		experimentVersion,
	)
	if err != nil {
		return "", err
	}

	return runId, nil
}

func (kfpp *KfpProvider) DeleteRun(_ string) error {
	// Not implemented for KFP provider
	// Required to satisfy the `Provider` interface
	return nil
}

func (kfpp *KfpProvider) CreateRunSchedule(
	rsd baseResource.RunScheduleDefinition,
) (string, error) {
	pipelineName, err := util.ResourceNameFromNamespacedName(rsd.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := kfpp.pipelineService.PipelineIdForName(pipelineName)
	if err != nil {
		return "", err
	}

	pipelineVersionId, err := kfpp.pipelineService.PipelineVersionIdForName(
		rsd.PipelineVersion,
		pipelineId,
	)
	if err != nil {
		return "", err
	}

	experimentVersion, err := kfpp.experimentService.ExperimentIdByName(rsd.ExperimentName)
	if err != nil {
		return "", err
	}

	jobId, err := kfpp.jobService.CreateJob(
		rsd,
		pipelineId,
		pipelineVersionId,
		experimentVersion,
	)
	if err != nil {
		return "", err
	}

	return jobId, nil
}

func (kfpp *KfpProvider) UpdateRunSchedule(
	rsd baseResource.RunScheduleDefinition,
	id string,
) (string, error) {
	if err := kfpp.DeleteRunSchedule(id); err != nil {
		return id, err
	}

	return kfpp.CreateRunSchedule(rsd)
}

func (kfpp *KfpProvider) DeleteRunSchedule(id string) error {
	return kfpp.jobService.DeleteJob(id)
}

func (kfpp *KfpProvider) CreateExperiment(
	ed baseResource.ExperimentDefinition,
) (string, error) {
	expId, err := kfpp.experimentService.CreateExperiment(
		ed.Name,
		ed.Description,
	)
	if err != nil {
		return "", err
	}

	return expId, nil
}

func (kfpp *KfpProvider) UpdateExperiment(
	ed baseResource.ExperimentDefinition,
	id string,
) (string, error) {
	if err := kfpp.DeleteExperiment(id); err != nil {
		return id, err
	}

	return kfpp.CreateExperiment(ed)
}

func (kfpp *KfpProvider) DeleteExperiment(id string) error {
	return kfpp.experimentService.DeleteExperiment(id)
}
