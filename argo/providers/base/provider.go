package base

import (
	"context"
	"github.com/argoproj/argo-events/eventsources/sources/generic"
	"github.com/sky-uk/kfp-operator/apis"
	pipelines "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
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

type RunScheduleDefinition struct {
	Name                 common.NamespacedName      `yaml:"name"`
	Version              string                     `yaml:"version"`
	PipelineName         common.NamespacedName      `yaml:"pipelineName"`
	PipelineVersion      string                     `yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName      `yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName      `yaml:"experimentName"`
	Schedule             string                     `yaml:"schedule"`
	RuntimeParameters    map[string]string          `yaml:"runtimeParameters"`
	Artifacts            []pipelines.OutputArtifact `yaml:"artifacts,omitempty"`
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

type Output struct {
	Id            string `yaml:"id"`
	ProviderError string `yaml:"providerError"`
}

type Provider[Config any] interface {
	CreatePipeline(ctx context.Context, providerConfig Config, pipelineDefinition PipelineDefinition, pipelineFilePath string) (string, error)
	UpdatePipeline(ctx context.Context, providerConfig Config, pipelineDefinition PipelineDefinition, id string, pipelineFilePath string) (string, error)
	DeletePipeline(ctx context.Context, providerConfig Config, id string) error

	CreateRun(ctx context.Context, providerConfig Config, runConfigurationDefinition RunDefinition) (string, error)
	DeleteRun(ctx context.Context, providerConfig Config, id string) error

	CreateRunSchedule(ctx context.Context, providerConfig Config, runScheduleDefinition RunScheduleDefinition) (string, error)
	UpdateRunSchedule(ctx context.Context, providerConfig Config, runScheduleDefinition RunScheduleDefinition, id string) (string, error)
	DeleteRunSchedule(ctx context.Context, providerConfig Config, id string) error

	CreateExperiment(ctx context.Context, providerConfig Config, experimentDefinition ExperimentDefinition) (string, error)
	UpdateExperiment(ctx context.Context, providerConfig Config, experimentDefinition ExperimentDefinition, id string) (string, error)
	DeleteExperiment(ctx context.Context, providerConfig Config, id string) error

	EventingServer(ctx context.Context, providerConfig Config) (generic.EventingServer, error)
}
