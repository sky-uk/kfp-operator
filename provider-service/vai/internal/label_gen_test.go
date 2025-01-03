//go:build unit

package internal

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Describe("DefaultLabelGen", func() {
	var lg = DefaultLabelGen{}

	Context("GenerateLabels", func() {
		When("value is RunDefinition", func() {
			It("should not error", func() {
				rd := randomBasicRunDefinition()
				_, err := lg.GenerateLabels(rd)

				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("value is RunScheduleDefinition", func() {
			It("should not error", func() {
				rsd := randomRunScheduleDefinition()
				_, err := lg.GenerateLabels(rsd)

				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("value is not RunDefinition or RunScheduleDefinition", func() {
			It("should return error", func() {
				_, err := lg.GenerateLabels(0)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("runLabelsFromRunDefinition", func() {
		When("RunConfigurationName and RunName is present", func() {
			It("generates run labels with RunConfigurationName and RunName", func() {
				rd := randomBasicRunDefinition()
				rl := lg.runLabelsFromRunDefinition(rd)

				Expect(rl[labels.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[labels.RunConfigurationName]).To(Equal(rd.RunConfigurationName.Name))
				Expect(rl[labels.RunConfigurationNamespace]).To(Equal(rd.RunConfigurationName.Namespace))
				Expect(rl[labels.RunName]).To(Equal(rd.Name.Name))
				Expect(rl[labels.RunNamespace]).To(Equal(rd.Name.Namespace))
			})
		})
		When("RunConfigurationName is empty", func() {
			It("generates run labels with RunName", func() {
				rd := randomBasicRunDefinition()
				rd.RunConfigurationName = common.NamespacedName{}
				rl := lg.runLabelsFromRunDefinition(rd)

				Expect(rl[labels.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[labels.RunName]).To(Equal(rd.Name.Name))
				Expect(rl[labels.RunNamespace]).To(Equal(rd.Name.Namespace))
				Expect(rl).NotTo(HaveKey(labels.RunConfigurationName))
				Expect(rl).NotTo(HaveKey(labels.RunConfigurationNamespace))
			})
		})
		When("RunName is empty", func() {
			It("generates run labels with RunName", func() {
				rd := randomBasicRunDefinition()
				rd.Name = common.NamespacedName{}
				rl := lg.runLabelsFromRunDefinition(rd)

				Expect(rl[labels.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[labels.RunConfigurationName]).To(Equal(rd.RunConfigurationName.Name))
				Expect(rl[labels.RunConfigurationNamespace]).To(Equal(rd.RunConfigurationName.Namespace))
				Expect(rl).NotTo(HaveKey(labels.RunName))
				Expect(rl).NotTo(HaveKey(labels.RunNamespace))
			})
		})
		It("replaces fullstops with dashes in pipelineVersion", func() {
			rd := randomBasicRunDefinition()
			rd.PipelineVersion = "0.4.0"
			rl := lg.runLabelsFromRunDefinition(rd)

			Expect(rl[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

	Context("runLabelsFromSchedule", func() {
		When("RunConfigurationName is present", func() {
			It("generates run labels with RunConfiguration name and namespace", func() {
				rsd := randomRunScheduleDefinition()
				rl := lg.runLabelsFromSchedule(rsd)

				Expect(rl[labels.PipelineName]).To(Equal(rsd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rsd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rsd.PipelineVersion))
				Expect(rl[labels.RunConfigurationName]).To(Equal(rsd.RunConfigurationName.Name))
				Expect(rl[labels.RunConfigurationNamespace]).To(Equal(rsd.RunConfigurationName.Namespace))
			})
		})
		When("RunConfigurationName is empty", func() {
			It("generates run labels without RunConfiguration name and namespace", func() {
				rsd := randomRunScheduleDefinition()
				rsd.RunConfigurationName = common.NamespacedName{}
				rl := lg.runLabelsFromSchedule(rsd)

				Expect(rl[labels.PipelineName]).To(Equal(rsd.PipelineName.Name))
				Expect(rl[labels.PipelineNamespace]).To(Equal(rsd.PipelineName.Namespace))
				Expect(rl[labels.PipelineVersion]).To(Equal(rsd.PipelineVersion))
				Expect(rl).NotTo(HaveKey(labels.RunConfigurationName))
				Expect(rl).NotTo(HaveKey(labels.RunConfigurationNamespace))
			})
		})

		It("replaces fullstops with dashes in pipelineVersion", func() {
			rd := randomBasicRunDefinition()
			rd.PipelineVersion = "0.4.0"
			rl := lg.runLabelsFromRunDefinition(rd)

			Expect(rl[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})
})
