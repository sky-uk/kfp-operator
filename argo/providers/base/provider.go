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

type Provider[Config any] interface {
	CreatePipeline(providerConfig Config, pipelineDefinition PipelineDefinition, pipelineFile string, ctx context.Context) (string, error)
	UpdatePipeline(providerConfig Config, pipelineDefinition PipelineDefinition, id string, pipelineFile string, ctx context.Context) (string, error)
	DeletePipeline(providerConfig Config, pipelineDefinition PipelineDefinition, id string, ctx context.Context) error
}
