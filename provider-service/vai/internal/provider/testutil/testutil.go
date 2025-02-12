package testutil

import (
	"encoding/json"

	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	commonTestUtil "github.com/sky-uk/kfp-operator/common/testutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RandomBasicRunDefinition() resource.RunDefinition {
	return resource.RunDefinition{
		Name:                 commonTestUtil.RandomNamespacedName(),
		PipelineName:         commonTestUtil.RandomNamespacedName(),
		PipelineVersion:      commonTestUtil.RandomString(),
		RunConfigurationName: commonTestUtil.RandomNamespacedName(),
	}
}

var Now = metav1.Now()

func RandomRunScheduleDefinition() resource.RunScheduleDefinition {
	return resource.RunScheduleDefinition{
		Name:                 commonTestUtil.RandomNamespacedName(),
		Version:              commonTestUtil.RandomString(),
		PipelineName:         commonTestUtil.RandomNamespacedName(),
		PipelineVersion:      commonTestUtil.RandomString(),
		RunConfigurationName: commonTestUtil.RandomNamespacedName(),
		ExperimentName:       commonTestUtil.RandomNamespacedName(),
		Schedule: pipelinesv1.Schedule{
			CronExpression: "1 1 0 0 0",
			StartTime:      &Now,
			EndTime:        &Now,
		},
	}
}

func RandomPipelineDefinition() resource.PipelineDefinition {
	return resource.PipelineDefinition{
		Name:          commonTestUtil.RandomNamespacedName(),
		Version:       commonTestUtil.RandomString(),
		Image:         commonTestUtil.RandomString(),
		TfxComponents: commonTestUtil.RandomString(),
		Env:           make([]apis.NamedValue, 0),
		BeamArgs:      make([]apis.NamedValue, 0),
	}
}

func RandomPipelineDefinitionWrapper() resource.PipelineDefinitionWrapper {
	return resource.PipelineDefinitionWrapper{
		PipelineDefinition: RandomPipelineDefinition(),
		CompiledPipeline:   json.RawMessage{},
	}
}
