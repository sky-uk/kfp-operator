package resources

import (
	"context"
	"github.com/sky-uk/kfp-operator/apis"
	pipelines "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
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

type Provider interface {
	CreatePipeline(ctx context.Context, pd PipelineDefinition, manifest map[string]interface{}) (string, error)
	UpdatePipeline(ctx context.Context, pd PipelineDefinition, id string, manifest map[string]interface{}) (string, error)
	DeletePipeline(ctx context.Context, id string) error

	CreateRun(ctx context.Context, rcd RunDefinition) (string, error)
	DeleteRun(ctx context.Context, id string) error

	CreateRunSchedule(ctx context.Context, rsd RunScheduleDefinition) (string, error)
	UpdateRunSchedule(ctx context.Context, rsd RunScheduleDefinition, id string) (string, error)
	DeleteRunSchedule(ctx context.Context, id string) error

	CreateExperiment(ctx context.Context, ed ExperimentDefinition) (string, error)
	UpdateExperiment(ctx context.Context, ed ExperimentDefinition, id string) (string, error)
	DeleteExperiment(ctx context.Context, id string) error
}
