package resource

import (
	"encoding/json"
	"github.com/sky-uk/kfp-operator/apis"
	pipelines "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/common/types"
)

type PipelineDefinition struct {
	Name          types.NamespacedName `json:"name" yaml:"name"`
	Version       string               `json:"version" yaml:"version"`
	Image         string               `json:"image" yaml:"image"`
	TfxComponents string               `json:"tfxComponents" yaml:"tfxComponents"`
	Env           []apis.NamedValue    `json:"env" yaml:"env"`
	BeamArgs      []apis.NamedValue    `json:"beamArgs" yaml:"beamArgs"`
}

// CompiledManifest represents the output of the python compile step, and
// describes what vertex ai or kubeflow pipelines should do.
type PipelineDefinitionWrapper struct {
	PipelineDefinition PipelineDefinition `json:"pipelineDefinition"`
	CompiledPipeline   json.RawMessage    `json:"compiledPipeline,omitempty"`
}

type ExperimentDefinition struct {
	Name        types.NamespacedName `yaml:"name"`
	Version     string               `yaml:"version"`
	Description string               `yaml:"description"`
}

type RunDefinition struct {
	Name                 types.NamespacedName       `yaml:"name"`
	Version              string                     `yaml:"version"`
	PipelineName         types.NamespacedName       `yaml:"pipelineName"`
	PipelineVersion      string                     `yaml:"pipelineVersion"`
	RunConfigurationName types.NamespacedName       `yaml:"runConfigurationName"`
	ExperimentName       types.NamespacedName       `yaml:"experimentName"`
	RuntimeParameters    map[string]string          `yaml:"runtimeParameters"`
	Artifacts            []pipelines.OutputArtifact `yaml:"artifacts,omitempty"`
}

type RunScheduleDefinition struct {
	Name                 types.NamespacedName       `yaml:"name"`
	Version              string                     `yaml:"version"`
	PipelineName         types.NamespacedName       `yaml:"pipelineName"`
	PipelineVersion      string                     `yaml:"pipelineVersion"`
	RunConfigurationName types.NamespacedName       `yaml:"runConfigurationName"`
	ExperimentName       types.NamespacedName       `yaml:"experimentName"`
	Schedule             pipelines.Schedule         `yaml:"schedule"`
	RuntimeParameters    map[string]string          `yaml:"runtimeParameters"`
	Artifacts            []pipelines.OutputArtifact `yaml:"artifacts,omitempty"`
}

type Provider interface {
	PipelineProvider
	RunProvider
	RunScheduleProvider
	ExperimentProvider
}

type PipelineProvider interface {
	CreatePipeline(pd PipelineDefinitionWrapper) (string, error)
	UpdatePipeline(pd PipelineDefinitionWrapper, id string) (string, error)
	DeletePipeline(id string) error
}

type RunProvider interface {
	CreateRun(rd RunDefinition) (string, error)
	DeleteRun(id string) error
}

type RunScheduleProvider interface {
	CreateRunSchedule(rsd RunScheduleDefinition) (string, error)
	UpdateRunSchedule(rsd RunScheduleDefinition, id string) (string, error)
	DeleteRunSchedule(id string) error
}

type ExperimentProvider interface {
	CreateExperiment(ed ExperimentDefinition) (string, error)
	UpdateExperiment(ed ExperimentDefinition, id string) (string, error)
	DeleteExperiment(id string) error
}

type UserError struct {
	E error
}

func (e *UserError) Error() string {
	return e.E.Error()
}
