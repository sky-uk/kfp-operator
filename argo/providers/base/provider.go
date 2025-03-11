package base

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	pipelines "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type PipelineDefinition struct {
	Name          common.NamespacedName `json:"name" yaml:"name"`
	Version       string                `json:"version" yaml:"version"`
	Image         string                `json:"image" yaml:"image"`
	TfxComponents string                `json:"tfxComponents" yaml:"tfxComponents"`
	Env           []apis.NamedValue     `json:"env" yaml:"env"`
	BeamArgs      []apis.NamedValue     `json:"beamArgs" yaml:"beamArgs"`
	Framework     string                `json:"framework" yaml:"framework"`
}

type ExperimentDefinition struct {
	Name        common.NamespacedName `json:"name" yaml:"name"`
	Version     string                `json:"version" yaml:"version"`
	Description string                `json:"description" yaml:"description"`
}

type RunScheduleDefinition struct {
	Name                 common.NamespacedName      `json:"name" yaml:"name"`
	Version              string                     `json:"version" yaml:"version"`
	PipelineName         common.NamespacedName      `json:"pipelineName" yaml:"pipelineName"`
	PipelineVersion      string                     `json:"pipelineVersion" yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName      `json:"runConfigurationName" yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName      `json:"experimentName" yaml:"experimentName"`
	Schedule             pipelines.Schedule         `json:"schedule" yaml:"schedule"`
	RuntimeParameters    map[string]string          `json:"runtimeParameters" yaml:"runtimeParameters"`
	Artifacts            []pipelines.OutputArtifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
}

type RunDefinition struct {
	Name                 common.NamespacedName      `json:"name" yaml:"name"`
	Version              string                     `json:"version" yaml:"version"`
	PipelineName         common.NamespacedName      `json:"pipelineName" yaml:"pipelineName"`
	PipelineVersion      string                     `json:"pipelineVersion" yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName      `json:"runConfigurationName" yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName      `json:"experimentName" yaml:"experimentName"`
	RuntimeParameters    map[string]string          `json:"runtimeParameters" yaml:"runtimeParameters"`
	Artifacts            []pipelines.OutputArtifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
}

type Output struct {
	Id            string `json:"id,omitempty" yaml:"id"`
	ProviderError string `json:"providerError,omitempty" yaml:"providerError"`
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
}
