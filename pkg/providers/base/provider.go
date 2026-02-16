package base

import (
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/pkg/common/triggers"
)

type PipelineDefinition struct {
	Name      common.NamespacedName          `json:"name" yaml:"name"`
	Version   string                         `json:"version" yaml:"version"`
	Image     string                         `json:"image" yaml:"image"`
	Env       []apis.NamedValue              `json:"env,omitempty" yaml:"env,omitempty"`
	Framework pipelineshub.PipelineFramework `json:"framework" yaml:"framework"`
}

type ExperimentDefinition struct {
	Name        common.NamespacedName `json:"name" yaml:"name"`
	Version     string                `json:"version" yaml:"version"`
	Description string                `json:"description" yaml:"description"`
}

type RunScheduleDefinition struct {
	Name                 common.NamespacedName         `json:"name" yaml:"name"`
	Version              string                        `json:"version" yaml:"version"`
	PipelineName         common.NamespacedName         `json:"pipelineName" yaml:"pipelineName"`
	PipelineVersion      string                        `json:"pipelineVersion" yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName         `json:"runConfigurationName" yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName         `json:"experimentName" yaml:"experimentName"`
	Schedule             pipelineshub.Schedule         `json:"schedule" yaml:"schedule"`
	Parameters           map[string]string             `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Artifacts            []pipelineshub.OutputArtifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
	TriggerIndicator     triggers.Indicator            `json:"triggerIndicator" yaml:"labels,omitempty"`
}

type RunDefinition struct {
	Name                 common.NamespacedName         `json:"name" yaml:"name"`
	Version              string                        `json:"version" yaml:"version"`
	PipelineName         common.NamespacedName         `json:"pipelineName" yaml:"pipelineName"`
	PipelineVersion      string                        `json:"pipelineVersion" yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName         `json:"runConfigurationName" yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName         `json:"experimentName" yaml:"experimentName"`
	Parameters           map[string]string             `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Artifacts            []pipelineshub.OutputArtifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
	TriggerIndicator     *triggers.Indicator           `json:"triggerIndicator,omitempty" yaml:"labels,omitempty"`
}

type Output struct {
	Id            string `json:"id,omitempty" yaml:"id"`
	ProviderError string `json:"providerError,omitempty" yaml:"providerError"`
}
