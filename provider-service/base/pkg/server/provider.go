package server

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
	pipelines "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/apis"
)

type PipelineDefinition struct {
	Name          common.NamespacedName `yaml:"name"`
	Version       string                `yaml:"version"`
	Image         string                `yaml:"image"`
	TfxComponents string                `yaml:"tfxComponents"`
	Env           []apis.NamedValue     `yaml:"env"`
	BeamArgs      []apis.NamedValue     `yaml:"beamArgs"`
}

type ExperimentDefinition struct {
	Name        common.NamespacedName `yaml:"name"`
	Version     string                `yaml:"version"`
	Description string                `yaml:"description"`
}

type RunDefinition struct {
	Name                 common.NamespacedName      `yaml:"name"`
	Version              string                     `yaml:"version"`
	PipelineName         common.NamespacedName      `yaml:"pipelineName"`
	PipelineVersion      string                     `yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName      `yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName      `yaml:"experimentName"`
	RuntimeParameters    map[string]string          `yaml:"runtimeParameters"`
	Artifacts            []pipelines.OutputArtifact `yaml:"artifacts,omitempty"`
}

type RunScheduleDefinition struct {
	Name                 common.NamespacedName      `yaml:"name"`
	Version              string                     `yaml:"version"`
	PipelineName         common.NamespacedName      `yaml:"pipelineName"`
	PipelineVersion      string                     `yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName      `yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName      `yaml:"experimentName"`
	Schedule             pipelines.Schedule         `yaml:"schedule"`
	RuntimeParameters    map[string]string          `yaml:"runtimeParameters"`
	Artifacts            []pipelines.OutputArtifact `yaml:"artifacts,omitempty"`
}

type Provider[Config any] interface {
	CreatePipeline(ctx context.Context, cfg Config, pd PipelineDefinition, pipelineFilePath string) (string, error)
	UpdatePipeline(ctx context.Context, cfg Config, pd PipelineDefinition, id string, pipelineFilePath string) (string, error)
	DeletePipeline(ctx context.Context, cfg Config, id string) error

	CreateRun(ctx context.Context, cfg Config, rcd RunDefinition) (string, error)
	DeleteRun(ctx context.Context, cfg Config, id string) error

	CreateRunSchedule(ctx context.Context, cfg Config, rsd RunScheduleDefinition) (string, error)
	UpdateRunSchedule(ctx context.Context, cfg Config, rsd RunScheduleDefinition, id string) (string, error)
	DeleteRunSchedule(ctx context.Context, cfg Config, id string) error

	CreateExperiment(ctx context.Context, cfg Config, ed ExperimentDefinition) (string, error)
	UpdateExperiment(ctx context.Context, cfg Config, ed ExperimentDefinition, id string) (string, error)
	DeleteExperiment(ctx context.Context, cfg Config, id string) error
}
