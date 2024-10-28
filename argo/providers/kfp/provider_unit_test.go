//go:build unit

package kfp

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/argo/providers/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

var now = metav1.Now()

func randomRunScheduleDefinition() RunScheduleDefinition {
	return RunScheduleDefinition{
		Name:                 common.RandomNamespacedName(),
		Version:              common.RandomString(),
		PipelineName:         common.RandomNamespacedName(),
		PipelineVersion:      common.RandomString(),
		RunConfigurationName: common.RandomNamespacedName(),
		ExperimentName:       common.RandomNamespacedName(),
		Schedule: pipelinesv1.Schedule{
			CronExpression: "0 1 1 0 0 0",
			StartTime:      &now,
			EndTime:        &now,
		},
	}
}

var _ = Context("KFP Provider", func() {
	Describe("createAPICronSchedule", func() {
		It("returns APICronScheudle with fields all set as expected", func() {
			runScheduleDefinition := randomRunScheduleDefinition()

			result, err := createAPICronSchedule(runScheduleDefinition)

			Expect(err).NotTo(HaveOccurred())
			Expect(result.Cron).To(Equal("0 1 1 0 0 0"))
			Expect(time.Time(result.StartTime)).To(Equal(now.Time))
			Expect(time.Time(result.EndTime)).To(Equal(now.Time))
		})
	})
})
