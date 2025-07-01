package provider

import (
	"context"
	"errors"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
)

type KfpProvider struct{}

func NewKfpProvider(config config.Config) (*KfpProvider, error) {
	return &KfpProvider{}, nil
}

var _ resource.Provider = &KfpProvider{}

func (*KfpProvider) CreatePipeline(
	ctx context.Context,
	ppd resource.PipelineDefinitionWrapper,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) UpdatePipeline(
	ctx context.Context,
	ppd resource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) DeletePipeline(
	ctx context.Context,
	id string,
) error {
	return errors.New("not implemented")
}

func (*KfpProvider) CreateRun(
	ctx context.Context,
	rd resource.RunDefinition,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) DeleteRun(
	ctx context.Context,
	id string,
) error {
	return errors.New("not implemented")
}

func (*KfpProvider) CreateRunSchedule(
	ctx context.Context,
	rsd resource.RunScheduleDefinition,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) UpdateRunSchedule(
	ctx context.Context,
	rsd resource.RunScheduleDefinition,
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

func (*KfpProvider) CreateExperiment(
	ctx context.Context,
	ed resource.ExperimentDefinition,
) (string, error) {
	return "", errors.New("not implemented")
}

func (*KfpProvider) UpdateExperiment(
	ctx context.Context,
	ed resource.ExperimentDefinition,
	id string,
) (string, error) {
	return "", errors.New("not implemented")
}

func (KfpProvider) DeleteExperiment(
	ctx context.Context,
	id string,
) error {
	return errors.New("not implemented")
}
