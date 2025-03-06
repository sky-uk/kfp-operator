package resource

import (
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type References struct {
	PipelineName         common.NamespacedName         `yaml:"pipelineName"`
	RunConfigurationName common.NamespacedName         `yaml:"runConfigurationName"`
	RunName              common.NamespacedName         `yaml:"runName"`
	Artifacts            []pipelineshub.OutputArtifact `yaml:"artifacts,omitempty"`
}
