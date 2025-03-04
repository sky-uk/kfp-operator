package resource

import (
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type References struct {
	PipelineName         common.NamespacedName        `yaml:"pipelineName"`
	RunConfigurationName common.NamespacedName        `yaml:"runConfigurationName"`
	RunName              common.NamespacedName        `yaml:"runName"`
	Artifacts            []pipelinesv1.OutputArtifact `yaml:"artifacts,omitempty"`
}
