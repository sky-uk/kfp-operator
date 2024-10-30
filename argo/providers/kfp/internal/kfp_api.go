package internal

import (
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var kfpApiConstants = struct {
	KfpResourceNotFoundCode int32
}{
	KfpResourceNotFoundCode: 5,
}

type ResourceReferences struct {
	PipelineName         common.NamespacedName        `yaml:"pipelineName"`
	RunConfigurationName common.NamespacedName        `yaml:"runConfigurationName"`
	RunName              common.NamespacedName        `yaml:"runName"`
	Artifacts            []pipelinesv1.OutputArtifact `yaml:"artifacts,omitempty"`
}
