//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/common"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/label"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/provider/testutil"
)

var _ = Describe("DefaultLabelGen", func() {
	var lg = DefaultLabelGen{}

	Context("GenerateLabels", func() {
		When("value is RunDefinition", func() {
			It("should not error", func() {
				rd := testutil.RandomBasicRunDefinition()
				_, err := lg.GenerateLabels(rd)

				Expect(err).ToNot(HaveOccurred())
			})
		})
		When("value is RunScheduleDefinition", func() {
			It("should not error", func() {
				rsd := testutil.RandomRunScheduleDefinition()
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
				rd := testutil.RandomBasicRunDefinition()
				rl := lg.runLabelsFromRunDefinition(rd)

				Expect(rl[label.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[label.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[label.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[label.RunConfigurationName]).To(Equal(rd.RunConfigurationName.Name))
				Expect(rl[label.RunConfigurationNamespace]).To(Equal(rd.RunConfigurationName.Namespace))
				Expect(rl[label.RunName]).To(Equal(rd.Name.Name))
				Expect(rl[label.RunNamespace]).To(Equal(rd.Name.Namespace))
			})
		})
		When("RunConfigurationName is empty", func() {
			It("generates run labels with RunName", func() {
				rd := testutil.RandomBasicRunDefinition()
				rd.RunConfigurationName = common.NamespacedName{}
				rl := lg.runLabelsFromRunDefinition(rd)

				Expect(rl[label.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[label.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[label.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[label.RunName]).To(Equal(rd.Name.Name))
				Expect(rl[label.RunNamespace]).To(Equal(rd.Name.Namespace))
				Expect(rl).NotTo(HaveKey(label.RunConfigurationName))
				Expect(rl).NotTo(HaveKey(label.RunConfigurationNamespace))
			})
		})
		When("RunName is empty", func() {
			It("generates run labels with RunName", func() {
				rd := testutil.RandomBasicRunDefinition()
				rd.Name = common.NamespacedName{}
				rl := lg.runLabelsFromRunDefinition(rd)

				Expect(rl[label.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[label.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[label.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[label.RunConfigurationName]).To(Equal(rd.RunConfigurationName.Name))
				Expect(rl[label.RunConfigurationNamespace]).To(Equal(rd.RunConfigurationName.Namespace))
				Expect(rl).NotTo(HaveKey(label.RunName))
				Expect(rl).NotTo(HaveKey(label.RunNamespace))
			})
		})
		It("replaces fullstops with dashes in pipelineVersion", func() {
			rd := testutil.RandomBasicRunDefinition()
			rd.PipelineVersion = "0.4.0"
			rl := lg.runLabelsFromRunDefinition(rd)

			Expect(rl[label.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

	Context("runLabelsFromSchedule", func() {
		When("RunConfigurationName is present", func() {
			It("generates run labels with RunConfiguration name and namespace", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				rl := lg.runLabelsFromSchedule(rsd)

				Expect(rl[label.PipelineName]).To(Equal(rsd.PipelineName.Name))
				Expect(rl[label.PipelineNamespace]).To(Equal(rsd.PipelineName.Namespace))
				Expect(rl[label.PipelineVersion]).To(Equal(rsd.PipelineVersion))
				Expect(rl[label.RunConfigurationName]).To(Equal(rsd.RunConfigurationName.Name))
				Expect(rl[label.RunConfigurationNamespace]).To(Equal(rsd.RunConfigurationName.Namespace))
			})
		})
		When("RunConfigurationName is empty", func() {
			It("generates run labels without RunConfiguration name and namespace", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				rsd.RunConfigurationName = common.NamespacedName{}
				rl := lg.runLabelsFromSchedule(rsd)

				Expect(rl[label.PipelineName]).To(Equal(rsd.PipelineName.Name))
				Expect(rl[label.PipelineNamespace]).To(Equal(rsd.PipelineName.Namespace))
				Expect(rl[label.PipelineVersion]).To(Equal(rsd.PipelineVersion))
				Expect(rl).NotTo(HaveKey(label.RunConfigurationName))
				Expect(rl).NotTo(HaveKey(label.RunConfigurationNamespace))
			})
		})

		It("replaces fullstops with dashes in pipelineVersion", func() {
			rd := testutil.RandomBasicRunDefinition()
			rd.PipelineVersion = "0.4.0"
			rl := lg.runLabelsFromRunDefinition(rd)

			Expect(rl[label.PipelineVersion]).To(Equal("0-4-0"))
		})
	})
})
