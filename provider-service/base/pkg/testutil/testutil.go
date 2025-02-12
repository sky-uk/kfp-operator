//go:build decoupled || unit

package testutil

import (
	"encoding/json"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/common/testutil"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var Now = metav1.Now()

func RandomRunScheduleDefinition() resource.RunScheduleDefinition {
	return resource.RunScheduleDefinition{
		Name:                 testutil.RandomNamespacedName(),
		Version:              testutil.RandomString(),
		PipelineName:         testutil.RandomNamespacedName(),
		PipelineVersion:      testutil.RandomString(),
		RunConfigurationName: testutil.RandomNamespacedName(),
		ExperimentName:       testutil.RandomNamespacedName(),
		Schedule: pipelinesv1.Schedule{
			CronExpression: "1 1 0 0 0",
			StartTime:      &Now,
			EndTime:        &Now,
		},
	}
}

func RandomPipelineDefinition() resource.PipelineDefinition {
	return resource.PipelineDefinition{
		Name:          testutil.RandomNamespacedName(),
		Version:       testutil.RandomString(),
		Image:         testutil.RandomString(),
		TfxComponents: testutil.RandomString(),
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

func RandomExperimentDefinition() resource.ExperimentDefinition {
	return resource.ExperimentDefinition{
		Name:        testutil.RandomNamespacedName(),
		Version:     testutil.RandomString(),
		Description: testutil.RandomString(),
	}
}

func RandomRunDefinition() resource.RunDefinition {
	return resource.RunDefinition{
		Name:                 testutil.RandomNamespacedName(),
		Version:              testutil.RandomString(),
		PipelineName:         testutil.RandomNamespacedName(),
		PipelineVersion:      testutil.RandomString(),
		RunConfigurationName: testutil.RandomNamespacedName(),
		ExperimentName:       testutil.RandomNamespacedName(),
	}
}
