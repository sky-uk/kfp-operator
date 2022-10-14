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

type Provider[Config any] interface {
	CreatePipeline(providerConfig Config, pipelineDefinition PipelineDefinition, pipelineFile string, ctx context.Context) (string, error)
	UpdatePipeline(providerConfig Config, pipelineDefinition PipelineDefinition, id string, pipelineFile string, ctx context.Context) (string, error)
	DeletePipeline(providerConfig Config, id string, ctx context.Context) error

	CreateRunConfiguration(providerConfig Config, runConfigurationDefinition RunConfigurationDefinition, ctx context.Context) (string, error)
	DeleteRunConfiguration(providerConfig Config, id string, ctx context.Context) error

	CreateExperiment(providerConfig Config, experimentDefinition ExperimentDefinition, ctx context.Context) (string, error)
	DeleteExperiment(providerConfig Config, id string, ctx context.Context) error
}
