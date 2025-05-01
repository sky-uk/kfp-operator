//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/label"
)

var _ = Describe("DefaultLabelGen", func() {
	var lg = DefaultLabelGen{
		providerName: common.NamespacedName{Name: "test-provider", Namespace: "test-namespace"},
	}

	Context("GenerateLabels", func() {
		When("value is not RunDefinition or RunScheduleDefinition", func() {
			It("should return error", func() {
				_, err := lg.GenerateLabels(0)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("GenerateLabels - runLabelsFromRunDefinition", func() {
		When("RunConfigurationName and RunName is present", func() {
			It("generates run labels with RunConfigurationName and RunName", func() {
				rd := testutil.RandomRunDefinition()
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl[label.ProviderName]).To(Equal(lg.providerName.Name))
				Expect(rl[label.ProviderNamespace]).To(Equal(lg.providerName.Namespace))
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
				rd := testutil.RandomRunDefinition()
				rd.RunConfigurationName = common.NamespacedName{}
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl[label.ProviderName]).To(Equal(lg.providerName.Name))
				Expect(rl[label.ProviderNamespace]).To(Equal(lg.providerName.Namespace))
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
				rd := testutil.RandomRunDefinition()
				rd.Name = common.NamespacedName{}
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl[label.ProviderName]).To(Equal(lg.providerName.Name))
				Expect(rl[label.ProviderNamespace]).To(Equal(lg.providerName.Namespace))
				Expect(rl[label.PipelineName]).To(Equal(rd.PipelineName.Name))
				Expect(rl[label.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
				Expect(rl[label.PipelineVersion]).To(Equal(rd.PipelineVersion))
				Expect(rl[label.RunConfigurationName]).To(Equal(rd.RunConfigurationName.Name))
				Expect(rl[label.RunConfigurationNamespace]).To(Equal(rd.RunConfigurationName.Namespace))
				Expect(rl).NotTo(HaveKey(label.RunName))
				Expect(rl).NotTo(HaveKey(label.RunNamespace))
			})
		})
	})

	Context("runLabelsFromRunDefinition", func() {
		It("replaces fullstops with dashes in pipelineVersion", func() {
			rd := testutil.RandomRunDefinition()
			rd.PipelineVersion = "0.4.0"
			rl := lg.runLabelsFromRunDefinition(rd)

			Expect(rl[label.PipelineVersion]).To(Equal("0-4-0"))
		})
	})

	Context("GenerateLabels - runLabelsFromSchedule", func() {
		When("RunConfigurationName is present", func() {
			It("generates run labels with RunConfiguration name and namespace", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				rl, err := lg.GenerateLabels(rsd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl[label.ProviderName]).To(Equal(lg.providerName.Name))
				Expect(rl[label.ProviderNamespace]).To(Equal(lg.providerName.Namespace))
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
				rl, err := lg.GenerateLabels(rsd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl[label.ProviderName]).To(Equal(lg.providerName.Name))
				Expect(rl[label.ProviderNamespace]).To(Equal(lg.providerName.Namespace))
				Expect(rl[label.PipelineName]).To(Equal(rsd.PipelineName.Name))
				Expect(rl[label.PipelineNamespace]).To(Equal(rsd.PipelineName.Namespace))
				Expect(rl[label.PipelineVersion]).To(Equal(rsd.PipelineVersion))
				Expect(rl).NotTo(HaveKey(label.RunConfigurationName))
				Expect(rl).NotTo(HaveKey(label.RunConfigurationNamespace))
			})
		})
	})

	Context("runLabelsFromSchedule", func() {
		It("replaces fullstops with dashes in pipelineVersion", func() {
			rd := testutil.RandomRunDefinition()
			rd.PipelineVersion = "0.4.0"
			rl := lg.runLabelsFromRunDefinition(rd)

			Expect(rl[label.PipelineVersion]).To(Equal("0-4-0"))
		})
	})
})
