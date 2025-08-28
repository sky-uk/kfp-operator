package resource

import (
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"time"
)

type References struct {
	PipelineName         common.NamespacedName         `yaml:"pipelineName"`
	RunConfigurationName common.NamespacedName         `yaml:"runConfigurationName"`
	RunName              common.NamespacedName         `yaml:"runName"`
	Artifacts            []pipelineshub.OutputArtifact `yaml:"artifacts,omitempty"`
	CreatedAt            *time.Time                    `yaml:"createdAt,omitempty"`
	FinishedAt           *time.Time                    `yaml:"finishedAt,omitempty"`
}
