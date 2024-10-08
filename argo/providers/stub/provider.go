package stub

import (
	"context"
	"errors"
	"fmt"
	"github.com/argoproj/argo-events/eventsources/sources/generic"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
)

type StubProviderConfig struct {
	Parameters Parameters `json:"parameters" yaml:"parameters"`
}

type Parameters struct {
	StubbedOutput base.Output   `json:"expectedOutput" yaml:"expectedOutput"`
	ExpectedInput ExpectedInput `json:"expectedInput" yaml:"expectedInput"`
}

type ResourceDefinition struct {
	Name    common.NamespacedName `yaml:"name"`
	Version string                `yaml:"version"`
}

type ExpectedInput struct {
	Id                 string             `yaml:"id"`
	ResourceDefinition ResourceDefinition `yaml:"resourceDefinition"`
}

type StubProvider struct {
}

func verifyResourceDefinition(providerConfig StubProviderConfig, actual ResourceDefinition) (string, error) {
	if providerConfig.Parameters.ExpectedInput.ResourceDefinition != actual {
		return "", fmt.Errorf("expected resource definition %+v did not match provided %+v", providerConfig.Parameters.ExpectedInput.ResourceDefinition, actual)
	}

	return providerConfig.Parameters.StubbedOutput.Id, errors.New(providerConfig.Parameters.StubbedOutput.ProviderError)
}

func verifyCreateCall(providerConfig StubProviderConfig, actual ResourceDefinition) (string, error) {
	return verifyResourceDefinition(providerConfig, actual)
}

func verifyUpdateCall(providerConfig StubProviderConfig, actual ResourceDefinition, id string) (string, error) {
	if providerConfig.Parameters.ExpectedInput.Id != id {
		return "", fmt.Errorf("expected id %s does not match provided id %s", providerConfig.Parameters.ExpectedInput, id)
	}

	return verifyResourceDefinition(providerConfig, actual)
}

func verifyDeleteCall(providerConfig StubProviderConfig, id string) error {
	if providerConfig.Parameters.ExpectedInput.Id != id {
		return fmt.Errorf("expected id %s does not match provided id %s", providerConfig.Parameters.ExpectedInput, id)
	}

	if providerConfig.Parameters.StubbedOutput.ProviderError != "" {
		return errors.New(providerConfig.Parameters.StubbedOutput.ProviderError)
	}

	return nil
}

func (s StubProvider) CreatePipeline(_ context.Context, providerConfig StubProviderConfig, resourceDefinition base.PipelineDefinition, _ string) (string, error) {
	return verifyCreateCall(providerConfig, ResourceDefinition{resourceDefinition.Name, resourceDefinition.Version})
}

func (s StubProvider) UpdatePipeline(_ context.Context, providerConfig StubProviderConfig, resourceDefinition base.PipelineDefinition, id string, _ string) (string, error) {
	return verifyUpdateCall(providerConfig, ResourceDefinition{resourceDefinition.Name, resourceDefinition.Version}, id)
}

func (s StubProvider) DeletePipeline(_ context.Context, providerConfig StubProviderConfig, id string) error {
	return verifyDeleteCall(providerConfig, id)
}

func (s StubProvider) CreateRun(_ context.Context, providerConfig StubProviderConfig, resourceDefinition base.RunDefinition) (string, error) {
	return verifyCreateCall(providerConfig, ResourceDefinition{resourceDefinition.Name, resourceDefinition.Version})
}

func (s StubProvider) DeleteRun(_ context.Context, providerConfig StubProviderConfig, id string) error {
	return verifyDeleteCall(providerConfig, id)
}

func (s StubProvider) CreateRunSchedule(_ context.Context, providerConfig StubProviderConfig, resourceDefinition base.RunScheduleDefinition) (string, error) {
	return verifyCreateCall(providerConfig, ResourceDefinition{Name: resourceDefinition.Name, Version: resourceDefinition.Version})
}

func (s StubProvider) UpdateRunSchedule(_ context.Context, providerConfig StubProviderConfig, resourceDefinition base.RunScheduleDefinition, id string) (string, error) {
	return verifyUpdateCall(providerConfig, ResourceDefinition{Name: resourceDefinition.Name, Version: resourceDefinition.Version}, id)
}

func (s StubProvider) DeleteRunSchedule(_ context.Context, providerConfig StubProviderConfig, id string) error {
	return verifyDeleteCall(providerConfig, id)
}

func (s StubProvider) CreateExperiment(_ context.Context, providerConfig StubProviderConfig, resourceDefinition base.ExperimentDefinition) (string, error) {
	return verifyCreateCall(providerConfig, ResourceDefinition{Name: resourceDefinition.Name, Version: resourceDefinition.Version})
}

func (s StubProvider) UpdateExperiment(_ context.Context, providerConfig StubProviderConfig, resourceDefinition base.ExperimentDefinition, id string) (string, error) {
	return verifyUpdateCall(providerConfig, ResourceDefinition{Name: resourceDefinition.Name, Version: resourceDefinition.Version}, id)
}

func (s StubProvider) DeleteExperiment(_ context.Context, providerConfig StubProviderConfig, id string) error {
	return verifyDeleteCall(providerConfig, id)
}

func (s StubProvider) EventingServer(_ context.Context, _ string, _ string) (generic.EventingServer, error) {
	panic("unimplemented")
}
