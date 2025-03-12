package base

import (
	"github.com/sky-uk/kfp-operator/apis"
	pipelines "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type PipelineDefinition struct {
	Name          common.NamespacedName `json:"name" yaml:"name"`
	Version       string                `json:"version" yaml:"version"`
	Image         string                `json:"image" yaml:"image"`
	TfxComponents string                `json:"tfxComponents" yaml:"tfxComponents"`
	Env           []apis.NamedValue     `json:"env,omitempty" yaml:"env,omitempty"`
	BeamArgs      []apis.NamedValue     `json:"beamArgs,omitempty" yaml:"beamArgs,omitempty"`
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
	RuntimeParameters    map[string]string          `json:"runtimeParameters,omitempty" yaml:"runtimeParameters,omitempty"`
	Artifacts            []pipelines.OutputArtifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
}

type RunDefinition struct {
	Name                 common.NamespacedName      `json:"name" yaml:"name"`
	Version              string                     `json:"version" yaml:"version"`
	PipelineName         common.NamespacedName      `json:"pipelineName" yaml:"pipelineName"`
	PipelineVersion      string                     `json:"pipelineVersion" yaml:"pipelineVersion"`
	RunConfigurationName common.NamespacedName      `json:"runConfigurationName" yaml:"runConfigurationName"`
	ExperimentName       common.NamespacedName      `json:"experimentName" yaml:"experimentName"`
	RuntimeParameters    map[string]string          `json:"runtimeParameters,omitempty" yaml:"runtimeParameters,omitempty"`
	Artifacts            []pipelines.OutputArtifact `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
}

type Output struct {
	Id            string `json:"id,omitempty" yaml:"id"`
	ProviderError string `json:"providerError,omitempty" yaml:"providerError"`
}
