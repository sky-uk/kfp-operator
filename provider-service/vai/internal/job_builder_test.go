//go:build unit

package internal

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func randomBasicRunDefinition() resource.RunDefinition {
	return resource.RunDefinition{
		Name:                 common.RandomNamespacedName(),
		PipelineName:         common.RandomNamespacedName(),
		PipelineVersion:      common.RandomString(),
		RunConfigurationName: common.RandomNamespacedName(),
	}
}

var now = metav1.Now()

func randomRunScheduleDefinition() resource.RunScheduleDefinition {
	return resource.RunScheduleDefinition{
		Name:                 common.RandomNamespacedName(),
		Version:              common.RandomString(),
		PipelineName:         common.RandomNamespacedName(),
		PipelineVersion:      common.RandomString(),
		RunConfigurationName: common.RandomNamespacedName(),
		ExperimentName:       common.RandomNamespacedName(),
		Schedule: pipelinesv1.Schedule{
			CronExpression: "1 1 0 0 0",
			StartTime:      &now,
			EndTime:        &now,
		},
	}
}

var _ = Describe("JobBuilder", func() {
	var jb = JobBuilder{}

	Context("runLabelsFromRunDefinition", func() {
		When("RunConfigurationName and RunName is present", func() {
			It("generates run labels with RunConfigurationName and RunName", func() {
				input := randomBasicRunDefinition()
				runLabels := jb.runLabelsFromRunDefinition(input)

				Expect(runLabels[labels.PipelineName]).To(Equal(input.PipelineName.Name))
				Expect(runLabels[labels.PipelineNamespace]).To(Equal(input.PipelineName.Namespace))
				Expect(runLabels[labels.PipelineVersion]).To(Equal(input.PipelineVersion))
				Expect(runLabels[labels.RunConfigurationName]).To(Equal(input.RunConfigurationName.Name))
				Expect(runLabels[labels.RunConfigurationNamespace]).To(Equal(input.RunConfigurationName.Namespace))
				Expect(runLabels[labels.RunName]).To(Equal(input.Name.Name))
				Expect(runLabels[labels.RunNamespace]).To(Equal(input.Name.Namespace))
			})
		})
		When("RunConfigurationName is empty", func() {
			It("generates run labels with RunName", func() {
				input := randomBasicRunDefinition()
				input.RunConfigurationName = common.NamespacedName{}
				runLabels := jb.runLabelsFromRunDefinition(input)

				Expect(runLabels[labels.PipelineName]).To(Equal(input.PipelineName.Name))
				Expect(runLabels[labels.PipelineNamespace]).To(Equal(input.PipelineName.Namespace))
				Expect(runLabels[labels.PipelineVersion]).To(Equal(input.PipelineVersion))
				Expect(runLabels[labels.RunName]).To(Equal(input.Name.Name))
				Expect(runLabels[labels.RunNamespace]).To(Equal(input.Name.Namespace))
				Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationName))
				Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationNamespace))
			})
		})
		When("RunName is empty", func() {
			It("generates run labels with RunName", func() {
				input := randomBasicRunDefinition()
				input.Name = common.NamespacedName{}
				runLabels := jb.runLabelsFromRunDefinition(input)

				Expect(runLabels[labels.PipelineName]).To(Equal(input.PipelineName.Name))
				Expect(runLabels[labels.PipelineNamespace]).To(Equal(input.PipelineName.Namespace))
				Expect(runLabels[labels.PipelineVersion]).To(Equal(input.PipelineVersion))
				Expect(runLabels[labels.RunConfigurationName]).To(Equal(input.RunConfigurationName.Name))
				Expect(runLabels[labels.RunConfigurationNamespace]).To(Equal(input.RunConfigurationName.Namespace))
				Expect(runLabels).NotTo(HaveKey(labels.RunName))
				Expect(runLabels).NotTo(HaveKey(labels.RunNamespace))
			})
		})
		It("replaces fullstops with dashes in pipelineVersion", func() {
			input := randomBasicRunDefinition()
			input.PipelineVersion = "0.4.0"
			runLabels := jb.runLabelsFromRunDefinition(input)

			Expect(runLabels[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

	Context("runLabelsFromSchedule", func() {
		When("RunConfigurationName is present", func() {
			It("generates run labels with RunConfiguration name and namespace", func() {
				input := randomRunScheduleDefinition()
				runLabels := jb.runLabelsFromSchedule(input)

				Expect(runLabels[labels.PipelineName]).To(Equal(input.PipelineName.Name))
				Expect(runLabels[labels.PipelineNamespace]).To(Equal(input.PipelineName.Namespace))
				Expect(runLabels[labels.PipelineVersion]).To(Equal(input.PipelineVersion))
				Expect(runLabels[labels.RunConfigurationName]).To(Equal(input.RunConfigurationName.Name))
				Expect(runLabels[labels.RunConfigurationNamespace]).To(Equal(input.RunConfigurationName.Namespace))
			})
		})
		When("RunConfigurationName is empty", func() {
			It("generates run labels without RunConfiguration name and namespace", func() {
				input := randomRunScheduleDefinition()
				input.RunConfigurationName = common.NamespacedName{}
				runLabels := jb.runLabelsFromSchedule(input)

				Expect(runLabels[labels.PipelineName]).To(Equal(input.PipelineName.Name))
				Expect(runLabels[labels.PipelineNamespace]).To(Equal(input.PipelineName.Namespace))
				Expect(runLabels[labels.PipelineVersion]).To(Equal(input.PipelineVersion))
				Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationName))
				Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationNamespace))
			})
		})
	})
})
