package base

import (
	"context"
)

type PipelineDefinition struct {
	RootLocation    string              `yaml:"rootLocation"`
	ServingLocation string              `yaml:"servingLocation"`
	Name            string              `yaml:"name"`
	Version         string              `yaml:"version"`
	Image           string              `yaml:"image"`
	TfxComponents   string              `yaml:"tfxComponents"`
	Env             map[string]string   `yaml:"env"`
	BeamArgs        map[string][]string `yaml:"beamArgs"`
}

type ExperimentDefinition struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

type RunConfigurationDefinition struct {
	Name            string `yaml:"name"`
	Version         string `yaml:"version"`
	PipelineName    string `yaml:"pipelineName"`
	PipelineVersion string `yaml:"pipelineVersion"`
	ExperimentName  string `yaml:"experimentName"`
	Schedule        string `yaml:"schedule"`
}

type Output struct {
	Id            string `yaml:"id"`
	ProviderError string `yaml:"providerError"`
}

type Provider[Config any] interface {
	CreatePipeline(ctx context.Context, providerConfig Config, pipelineDefinition PipelineDefinition, pipelineFile string) (string, error)
	UpdatePipeline(ctx context.Context, providerConfig Config, pipelineDefinition PipelineDefinition, id string, pipelineFile string) (string, error)
	DeletePipeline(ctx context.Context, providerConfig Config, id string) error

	CreateRunConfiguration(ctx context.Context, providerConfig Config, runConfigurationDefinition RunConfigurationDefinition) (string, error)
	UpdateRunConfiguration(ctx context.Context, providerConfig Config, runConfigurationDefinition RunConfigurationDefinition, id string) (string, error)
	DeleteRunConfiguration(ctx context.Context, providerConfig Config, id string) error

	CreateExperiment(ctx context.Context, providerConfig Config, experimentDefinition ExperimentDefinition) (string, error)
	UpdateExperiment(ctx context.Context, providerConfig Config, experimentDefinition ExperimentDefinition, id string) (string, error)
	DeleteExperiment(ctx context.Context, providerConfig Config, id string) error
}
