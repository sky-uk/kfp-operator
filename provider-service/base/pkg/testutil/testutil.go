//go:build decoupled || unit

package testutil

import (
	"encoding/json"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Now = metav1.Now()

func RandomRunScheduleDefinition() resource.RunScheduleDefinition {
	return resource.RunScheduleDefinition{
		Name:                 common.RandomNamespacedName(),
		Version:              common.RandomString(),
		PipelineName:         common.RandomNamespacedName(),
		PipelineVersion:      common.RandomString(),
		RunConfigurationName: common.RandomNamespacedName(),
		ExperimentName:       common.RandomNamespacedName(),
		Schedule: pipelineshub.Schedule{
			CronExpression: "1 1 0 0 0",
			StartTime:      &Now,
			EndTime:        &Now,
		},
	}
}

func RandomPipelineDefinition() resource.PipelineDefinition {
	return resource.PipelineDefinition{
		Name:      common.RandomNamespacedName(),
		Version:   common.RandomString(),
		Image:     common.RandomString(),
		Env:       make([]apis.NamedValue, 0),
		Framework: resource.PipelineFramework{Type: common.RandomString()},
	}
}

func RandomPipelineDefinitionWrapper() resource.PipelineDefinitionWrapper {
	return resource.PipelineDefinitionWrapper{
		PipelineDefinition: RandomPipelineDefinition(),
		CompiledPipeline:   json.RawMessage{},
	}
}

func RandomExperimentDefinition() resource.ExperimentDefinition {
	return resource.ExperimentDefinition{
		Name:        common.RandomNamespacedName(),
		Version:     common.RandomString(),
		Description: common.RandomString(),
	}
}

func RandomRunDefinition() resource.RunDefinition {
	return resource.RunDefinition{
		Name:                 common.RandomNamespacedName(),
		Version:              common.RandomString(),
		PipelineName:         common.RandomNamespacedName(),
		PipelineVersion:      common.RandomString(),
		RunConfigurationName: common.RandomNamespacedName(),
		ExperimentName:       common.RandomNamespacedName(),
	}
}
