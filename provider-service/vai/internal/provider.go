package internal

import (
	"context"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/file"
)

type VAIProvider struct {
	ctx         context.Context
	config      VAIProviderConfig
	fileHandler file.FileHandler
}

func NewProvider(ctx context.Context, config VAIProviderConfig) (*VAIProvider, error) {
	fh, err := file.NewGcsFileHandler(ctx, config.Parameters.GcsEndpoint)
	if err != nil {
		return nil, err
	}

	return &VAIProvider{
		ctx:         ctx,
		config:      config,
		fileHandler: &fh,
	}, nil
}

func (vaip *VAIProvider) CreatePipeline(pd resource.PipelineDefinition) (string, error) {
	pipelineId, err := vaip.UpdatePipeline(pd, "")
	if err != nil {
		return "", err
	}
	return pipelineId, nil
}

func (vaip *VAIProvider) UpdatePipeline(pd resource.PipelineDefinition, id string) (string, error) {
	pipelineId, err := pd.Name.String()
	if err != nil {
		return "", err
	}

	storageObject, err := vaip.config.pipelineStorageObject(pd.Name, pd.Version)
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

func (vaip *VAIProvider) CreateRun(rcd resource.RunDefinition) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) DeleteRun(id string) error {
	return nil
}

func (vaip *VAIProvider) CreateRunSchedule(rsd resource.RunScheduleDefinition) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) UpdateRunSchedule(rsd resource.RunScheduleDefinition, id string) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) DeleteRunSchedule(id string) error {
	return nil
}

func (vaip *VAIProvider) CreateExperiment(ed resource.ExperimentDefinition) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) UpdateExperiment(ed resource.ExperimentDefinition, id string) (string, error) {
	return "", nil
}

func (vaip *VAIProvider) DeleteExperiment(id string) error {
	return nil
}
