package stub

import (
	"context"
	"errors"
	"fmt"
	"github.com/sky-uk/kfp-operator/providers/base"
	"github.com/sky-uk/kfp-operator/providers/base/generic"
)

type StubProviderConfig struct {
	ExpectedOutput base.Output `yaml:"expectedOutput"`
	ExpectedId     string      `yaml:"expectedId"`
}

type StubProvider struct {
}

func (s StubProvider) CreatePipeline(ctx context.Context, providerConfig StubProviderConfig, pipelineDefinition base.PipelineDefinition, pipelineFile string) (string, error) {
	return providerConfig.ExpectedOutput.Id, errors.New(providerConfig.ExpectedOutput.ProviderError)
}

func (s StubProvider) UpdatePipeline(ctx context.Context, providerConfig StubProviderConfig, pipelineDefinition base.PipelineDefinition, id string, pipelineFile string) (string, error) {
	if providerConfig.ExpectedId != id {
		return "", fmt.Errorf("expected id %s does not match provided id %s", providerConfig.ExpectedId, id)
	}

	return providerConfig.ExpectedOutput.Id, errors.New(providerConfig.ExpectedOutput.ProviderError)
}

func (s StubProvider) DeletePipeline(ctx context.Context, providerConfig StubProviderConfig, id string) error {
	if providerConfig.ExpectedId != id {
		return fmt.Errorf("expected id %s does not match provided id %s", providerConfig.ExpectedId, id)
	}

	if providerConfig.ExpectedOutput.ProviderError != "" {
		return errors.New(providerConfig.ExpectedOutput.ProviderError)
	}

	return nil
}

func (s StubProvider) CreateRunConfiguration(ctx context.Context, providerConfig StubProviderConfig, runConfigurationDefinition base.RunConfigurationDefinition) (string, error) {
	return providerConfig.ExpectedOutput.Id, errors.New(providerConfig.ExpectedOutput.ProviderError)
}

func (s StubProvider) UpdateRunConfiguration(ctx context.Context, providerConfig StubProviderConfig, runConfigurationDefinition base.RunConfigurationDefinition, id string) (string, error) {
	if providerConfig.ExpectedId != id {
		return "", fmt.Errorf("expected id %s does not match provided id %s", providerConfig.ExpectedId, id)
	}

	return providerConfig.ExpectedOutput.Id, errors.New(providerConfig.ExpectedOutput.ProviderError)
}

func (s StubProvider) DeleteRunConfiguration(ctx context.Context, providerConfig StubProviderConfig, id string) error {
	if providerConfig.ExpectedId != id {
		return fmt.Errorf("expected id %s does not match provided id %s", providerConfig.ExpectedId, id)
	}

	if providerConfig.ExpectedOutput.ProviderError != "" {
		return errors.New(providerConfig.ExpectedOutput.ProviderError)
	}

	return nil
}

func (s StubProvider) CreateExperiment(_ context.Context, providerConfig StubProviderConfig, experimentDefinition base.ExperimentDefinition) (string, error) {
	return providerConfig.ExpectedOutput.Id, errors.New(providerConfig.ExpectedOutput.ProviderError)
}

func (s StubProvider) UpdateExperiment(_ context.Context, providerConfig StubProviderConfig, experimentDefinition base.ExperimentDefinition, id string) (string, error) {
	if providerConfig.ExpectedId != id {
		return "", fmt.Errorf("expected id %s does not match provided id %s", providerConfig.ExpectedId, id)
	}

	return providerConfig.ExpectedOutput.Id, errors.New(providerConfig.ExpectedOutput.ProviderError)
}

func (s StubProvider) DeleteExperiment(_ context.Context, providerConfig StubProviderConfig, id string) error {
	if providerConfig.ExpectedId != id {
		return fmt.Errorf("expected id %s does not match provided id %s", providerConfig.ExpectedId, id)
	}

	if providerConfig.ExpectedOutput.ProviderError != "" {
		return errors.New(providerConfig.ExpectedOutput.ProviderError)
	}

	return nil
}

func (s StubProvider) EventingServer(ctx context.Context, providerConfig StubProviderConfig) (generic.EventingServer, error) {
	//TODO implement me
	panic("implement me")
}
