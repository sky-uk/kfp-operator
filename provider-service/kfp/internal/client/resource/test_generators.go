//go:build decoupled || unit

package resource

import (
	"github.com/sky-uk/kfp-operator/pkg/common"
	"time"
)

func RandomReferences() References {
	// Round is used to remove monotonic clock from time.Now() to ensure that
	// the time is compatible with equality checks
	staticTime := time.Now().UTC().Round(0)
	return References{
		RunConfigurationName: common.RandomNamespacedName(),
		RunName:              common.RandomNamespacedName(),
		PipelineName:         common.RandomNamespacedName(),
		CreatedAt:            &staticTime,
		FinishedAt:           &staticTime,
	}
}
