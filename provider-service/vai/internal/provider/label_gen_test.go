//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/common/triggers"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/label"
)

var _ = Describe("DefaultLabelGen", func() {
	var lg = DefaultLabelGen{
		providerName: common.NamespacedName{Name: "test-provider", Namespace: "test-namespace"},
	}

	Context("GenerateLabels", func() {
		When("value is RunDefinition", func() {
			It("generates labels with provider name and namespace", func() {
				rd := testutil.RandomRunDefinition()
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl[label.ProviderName]).To(Equal(lg.providerName.Name))
				Expect(rl[label.ProviderNamespace]).To(Equal(lg.providerName.Namespace))
			})
		})

		When("value is RunScheduleDefinition", func() {
			It("generates labels with provider name and namespace", func() {
				rs := testutil.RandomRunScheduleDefinition()
				rl, err := lg.GenerateLabels(rs)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl[label.ProviderName]).To(Equal(lg.providerName.Name))
				Expect(rl[label.ProviderNamespace]).To(Equal(lg.providerName.Namespace))
			})
		})

		When("value is not RunDefinition or RunScheduleDefinition", func() {
			It("should return error", func() {
				_, err := lg.GenerateLabels(0)

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("GenerateLabels - runLabelsFromRunDefinition", func() {
		It("generates labels with pipeline name, namespace and version", func() {
			rd := testutil.RandomRunDefinition()
			rl, err := lg.GenerateLabels(rd)
			Expect(err).ToNot(HaveOccurred())
			Expect(rl[label.PipelineName]).To(Equal(rd.PipelineName.Name))
			Expect(rl[label.PipelineNamespace]).To(Equal(rd.PipelineName.Namespace))
			Expect(rl[label.PipelineVersion]).To(Equal(rd.PipelineVersion))
		})

		When("RunConfigurationName and RunName is present", func() {
			It("generates run labels with RunConfigurationName and RunName", func() {
				rd := testutil.RandomRunDefinition()
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

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

				Expect(rl[label.RunName]).To(Equal(rd.Name.Name))
				Expect(rl[label.RunNamespace]).To(Equal(rd.Name.Namespace))
				Expect(rl).NotTo(HaveKey(label.RunConfigurationName))
				Expect(rl).NotTo(HaveKey(label.RunConfigurationNamespace))
			})
		})

		When("RunName is empty", func() {
			It("generates run labels with RunConfiguration", func() {
				rd := testutil.RandomRunDefinition()
				rd.Name = common.NamespacedName{}
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl[label.RunConfigurationName]).To(Equal(rd.RunConfigurationName.Name))
				Expect(rl[label.RunConfigurationNamespace]).To(Equal(rd.RunConfigurationName.Namespace))
				Expect(rl).NotTo(HaveKey(label.RunName))
				Expect(rl).NotTo(HaveKey(label.RunNamespace))
			})
		})
	})

	Context("GenerateLabels - runLabelsFromSchedule", func() {
		It("generates labels with pipeline name, namespace and version", func() {
			rs := testutil.RandomRunScheduleDefinition()
			rl, err := lg.GenerateLabels(rs)
			Expect(err).ToNot(HaveOccurred())
			Expect(rl[label.PipelineName]).To(Equal(rs.PipelineName.Name))
			Expect(rl[label.PipelineNamespace]).To(Equal(rs.PipelineName.Namespace))
			Expect(rl[label.PipelineVersion]).To(Equal(rs.PipelineVersion))
		})

		When("RunConfigurationName is present", func() {
			It("generates run labels with RunConfiguration name and namespace", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				rl, err := lg.GenerateLabels(rsd)
				Expect(err).ToNot(HaveOccurred())

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

				Expect(rl).NotTo(HaveKey(label.RunConfigurationName))
				Expect(rl).NotTo(HaveKey(label.RunConfigurationNamespace))
			})
		})

		It("generates run labels with trigger type and source", func() {
			rsd := testutil.RandomRunScheduleDefinition()
			rl, err := lg.GenerateLabels(rsd)
			Expect(err).ToNot(HaveOccurred())

			Expect(rl[triggers.Type]).To(Equal(triggers.Schedule))
			Expect(rl[triggers.Source]).To(Equal(rsd.Name.Name))
			Expect(rl[triggers.SourceNamespace]).To(Equal(rsd.Name.Namespace))
		})
	})
})
