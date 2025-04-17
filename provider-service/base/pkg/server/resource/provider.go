package resource

import (
	"encoding/json"
	"fmt"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/sky-uk/kfp-operator/apis"
	pipelines "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type PipelineDefinition struct {
	Name      common.NamespacedName `json:"name" yaml:"name"`
	Version   string                `json:"version" yaml:"version"`
	Image     string                `json:"image" yaml:"image"`
	Env       []apis.NamedValue     `json:"env" yaml:"env"`
	Framework PipelineFramework     `json:"framework" yaml:"framework"`
}

type PipelineFramework struct {
	Type       string                           `json:"type" yaml:"type"`
	Parameters map[string]*apiextensionsv1.JSON `json:"parameters" yaml:"parameters"`
}

// CompiledManifest represents the output of the python compile step, and
// describes what vertex ai or kubeflow pipelines should do.
type PipelineDefinitionWrapper struct {
	PipelineDefinition PipelineDefinition `json:"pipelineDefinition"`
	CompiledPipeline   json.RawMessage    `json:"compiledPipeline,omitempty"`
}

type ExperimentDefinition struct {
	Name        common.NamespacedName `json:"name" yaml:"name"`
	Version     string                `json:"version" yaml:"version"`
	Description string                `json:"description" yaml:"description"`
}

type RunDefinition struct {
	Name                 common.NamespacedName      `json:"name" yaml:"name"`
	Version              string                     `json:"version" yaml:"version"`
	PipelineName         common.NamespacedName      `json:"pipelineName" yaml:"pipelineName"`
	PipelineVersion      string                     `json:"pipelineVersion" yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName      `json:"runConfigurationName" yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName      `json:"experimentName" yaml:"experimentName"`
	Parameters           map[string]string          `json:"parameters" yaml:"parameters"`
	Artifacts            []pipelines.OutputArtifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
}

type RunScheduleDefinition struct {
	Name                 common.NamespacedName      `json:"name" yaml:"name"`
	Version              string                     `json:"version" yaml:"version"`
	PipelineName         common.NamespacedName      `json:"pipelineName" yaml:"pipelineName"`
	PipelineVersion      string                     `json:"pipelineVersion" yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName      `json:"runConfigurationName" yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName      `json:"experimentName" yaml:"experimentName"`
	Schedule             pipelines.Schedule         `json:"schedule" yaml:"schedule"`
	Parameters           map[string]string          `json:"parameters" yaml:"parameters"`
	Artifacts            []pipelines.OutputArtifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
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

type UnimplementedError struct {
	Method       string
	ResourceType string
}

func (e *UnimplementedError) Error() string {
	return fmt.Sprintf("Method %s unimplemented for resource %s", e.Method, e.ResourceType)
}
