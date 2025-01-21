package testutil

import (
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"

	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
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
		Schedule: pipelinesv1.Schedule{
			CronExpression: "1 1 0 0 0",
			StartTime:      &Now,
			EndTime:        &Now,
		},
	}
}
