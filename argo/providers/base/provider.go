package base

import (
	"context"
)

type PipelineConfig struct {
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
	CreatePipeline(providerConfig Config, pipelineConfig PipelineConfig, pipelineFile string, ctx context.Context) (string, error)
	UpdatePipeline(providerConfig Config, pipelineConfig PipelineConfig, id string, pipelineFile string, ctx context.Context) (string, error)
	DeletePipeline(providerConfig Config, pipelineConfig PipelineConfig, id string, ctx context.Context) error
}
