package provider

import (
	"context"
	"errors"

	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/util"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type KfpProvider struct {
	config                *config.Config
	pipelineUploadService PipelineUploadService
	pipelineService       PipelineService
	runService            RunService
	experimentService     ExperimentService
}

func NewKfpProvider(config *config.Config) (*KfpProvider, error) {
	pipelineUploadService, err := NewPipelineUploadService(
		config.Parameters.RestKfpApiUrl,
	)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(
		config.Parameters.GrpcKfpApiAddress,
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

	return &KfpProvider{
		config:                config,
		pipelineUploadService: pipelineUploadService,
		pipelineService:       pipelineService,
		runService:            runService,
		experimentService:     experimentService,
	}, nil
}

var _ resource.Provider = &KfpProvider{}

func (p *KfpProvider) CreatePipeline(
	ctx context.Context,
	pdw resource.PipelineDefinitionWrapper,
) (string, error) {
	pipelineName, err := pdw.PipelineDefinition.Name.String()
	if err != nil {
		return "", err
	}

	pipelineId, err := p.pipelineUploadService.UploadPipeline(
		ctx,
		pdw.CompiledPipeline,
		pipelineName,
	)
	if err != nil {
		return "", err
	}
	return p.UpdatePipeline(ctx, pdw, pipelineId)
}

func (p *KfpProvider) UpdatePipeline(
	ctx context.Context,
	pdw resource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	if err := p.pipelineUploadService.UploadPipelineVersion(
		ctx,
		id,
		pdw.CompiledPipeline,
		pdw.PipelineDefinition.Version,
	); err != nil {
		return "", err
	}

	return id, nil
}

func (p *KfpProvider) DeletePipeline(
	ctx context.Context,
	id string,
) error {
	return p.pipelineService.DeletePipeline(ctx, id)
}

func (p *KfpProvider) CreateRun(
	ctx context.Context,
	rd base.RunDefinition,
) (string, error) {
	pipelineName, err := util.ResourceNameFromNamespacedName(rd.PipelineName)
	if err != nil {
		return "", err
	}

	pipelineId, err := p.pipelineService.PipelineIdForName(ctx, pipelineName)
	if err != nil {
		return "", err
	}

	pipelineVersionId, err := p.pipelineService.PipelineVersionIdForName(
		ctx,
		rd.PipelineVersion,
		pipelineId,
	)
	if err != nil {
		return "", err
	}

	experimentId, err := p.experimentService.ExperimentIdByName(ctx, rd.ExperimentName)
	if err != nil {
		return "", err
	}

	runId, err := p.runService.CreateRun(
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

func (*KfpProvider) DeleteRun(
	ctx context.Context,
	id string,
) error {
	return errors.New("not implemented")
}

func (*KfpProvider) CreateRunSchedule(
	ctx context.Context,
	rsd base.RunScheduleDefinition,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) UpdateRunSchedule(
	ctx context.Context,
	rsd base.RunScheduleDefinition,
	id string,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) DeleteRunSchedule(
	ctx context.Context,
	id string,
) error {
	return errors.New("not implemented")
}

func (p *KfpProvider) CreateExperiment(
	ctx context.Context,
	ed base.ExperimentDefinition,
) (string, error) {
	expId, err := p.experimentService.CreateExperiment(
		ctx,
		ed.Name,
		ed.Description,
	)
	if err != nil {
		return "", err
	}

	return expId, nil
}

func (p *KfpProvider) UpdateExperiment(
	ctx context.Context,
	ed base.ExperimentDefinition,
	id string,
) (string, error) {
	if err := p.DeleteExperiment(ctx, id); err != nil {
		return id, err
	}

	return p.CreateExperiment(ctx, ed)
}

func (p *KfpProvider) DeleteExperiment(
	ctx context.Context,
	id string,
) error {
	return p.experimentService.DeleteExperiment(ctx, id)
}
