package resource

import (
	"github.com/sky-uk/kfp-operator/pkg/common"
	"time"
)

type References struct {
	PipelineName         common.NamespacedName         `yaml:"pipelineName"`
	RunConfigurationName common.NamespacedName         `yaml:"runConfigurationName"`
	RunName              common.NamespacedName         `yaml:"runName"`
	Artifacts            map[string]map[string][]int64 `yaml:"artifacts,omitempty"`
	CreatedAt            *time.Time                    `yaml:"createdAt,omitempty"`
	FinishedAt           *time.Time                    `yaml:"finishedAt,omitempty"`
}
