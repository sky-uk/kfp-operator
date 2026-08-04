//go:build unit

package label

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/pkg/common/triggers"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/testutil"
)

var _ = Describe("DefaultLabelGen", func() {
	var lg = DefaultLabelGen{
		ProviderName: common.NamespacedName{Name: "test-provider", Namespace: "test-namespace"},
	}

	Context("GenerateLabels", func() {
		When("value is RunDefinition", func() {
			It("generates labels with provider name and namespace", func() {
				rd := testutil.RandomRunDefinition()
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl).To(HaveKeyWithValue(ProviderName, lg.ProviderName.Name))
				Expect(rl).To(HaveKeyWithValue(ProviderNamespace, lg.ProviderName.Namespace))
			})
		})

		When("value is RunScheduleDefinition", func() {
			It("generates labels with provider name and namespace", func() {
				rs := testutil.RandomRunScheduleDefinition()
				rl, err := lg.GenerateLabels(rs)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl).To(HaveKeyWithValue(ProviderName, lg.ProviderName.Name))
				Expect(rl).To(HaveKeyWithValue(ProviderNamespace, lg.ProviderName.Namespace))
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
			Expect(rl).To(HaveKeyWithValue(PipelineName, rd.PipelineName.Name))
			Expect(rl).To(HaveKeyWithValue(PipelineNamespace, rd.PipelineName.Namespace))
			Expect(rl).To(HaveKeyWithValue(PipelineVersion, rd.PipelineVersion))
		})

		When("RunConfigurationName and RunName is present", func() {
			It("generates run labels with RunConfigurationName and RunName", func() {
				rd := testutil.RandomRunDefinition()
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl).To(HaveKeyWithValue(RunConfigurationName, rd.RunConfigurationName.Name))
				Expect(rl).To(HaveKeyWithValue(RunConfigurationNamespace, rd.RunConfigurationName.Namespace))
				Expect(rl).To(HaveKeyWithValue(RunName, rd.Name.Name))
				Expect(rl).To(HaveKeyWithValue(RunNamespace, rd.Name.Namespace))
			})
		})

		When("RunConfigurationName is empty", func() {
			It("generates run labels with RunName", func() {
				rd := testutil.RandomRunDefinition()
				rd.RunConfigurationName = common.NamespacedName{}
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl).To(HaveKeyWithValue(RunName, rd.Name.Name))
				Expect(rl).To(HaveKeyWithValue(RunNamespace, rd.Name.Namespace))
				Expect(rl).NotTo(HaveKey(RunConfigurationName))
				Expect(rl).NotTo(HaveKey(RunConfigurationNamespace))
			})
		})

		When("RunName is empty", func() {
			It("generates run labels with RunConfiguration", func() {
				rd := testutil.RandomRunDefinition()
				rd.Name = common.NamespacedName{}
				rl, err := lg.GenerateLabels(rd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl).To(HaveKeyWithValue(RunConfigurationName, rd.RunConfigurationName.Name))
				Expect(rl).To(HaveKeyWithValue(RunConfigurationNamespace, rd.RunConfigurationName.Namespace))
				Expect(rl).NotTo(HaveKey(RunName))
				Expect(rl).NotTo(HaveKey(RunNamespace))
			})
		})
	})

	Context("GenerateLabels - runLabelsFromSchedule", func() {
		It("generates labels with pipeline name, namespace and version", func() {
			rs := testutil.RandomRunScheduleDefinition()
			rl, err := lg.GenerateLabels(rs)
			Expect(err).ToNot(HaveOccurred())
			Expect(rl).To(HaveKeyWithValue(PipelineName, rs.PipelineName.Name))
			Expect(rl).To(HaveKeyWithValue(PipelineNamespace, rs.PipelineName.Namespace))
			Expect(rl).To(HaveKeyWithValue(PipelineVersion, rs.PipelineVersion))
		})

		When("RunConfigurationName is present", func() {
			It("generates run labels with RunConfiguration name and namespace", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				rl, err := lg.GenerateLabels(rsd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl).To(HaveKeyWithValue(RunConfigurationName, rsd.RunConfigurationName.Name))
				Expect(rl).To(HaveKeyWithValue(RunConfigurationNamespace, rsd.RunConfigurationName.Namespace))
			})
		})

		When("RunConfigurationName is empty", func() {
			It("generates run labels without RunConfiguration name and namespace", func() {
				rsd := testutil.RandomRunScheduleDefinition()
				rsd.RunConfigurationName = common.NamespacedName{}
				rl, err := lg.GenerateLabels(rsd)
				Expect(err).ToNot(HaveOccurred())

				Expect(rl).NotTo(HaveKey(RunConfigurationName))
				Expect(rl).NotTo(HaveKey(RunConfigurationNamespace))
			})
		})

		It("generates run labels with trigger type and source", func() {
			rsd := testutil.RandomRunScheduleDefinition()
			rl, err := lg.GenerateLabels(rsd)
			Expect(err).ToNot(HaveOccurred())

			Expect(rl).To(HaveKeyWithValue(triggers.Type, triggers.Schedule))
			Expect(rl).To(HaveKeyWithValue(triggers.Source, rsd.Name.Name))
			Expect(rl).To(HaveKeyWithValue(triggers.SourceNamespace, rsd.Name.Namespace))
		})
	})
})
