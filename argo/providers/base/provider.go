package base

import (
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
	Framework     string                `yaml:"framework"`
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
	Schedule             pipelines.Schedule         `yaml:"schedule"`
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
	Id            string `json:"id,omitempty" yaml:"id"`
	ProviderError string `json:"providerError,omitempty" yaml:"providerError"`
}
