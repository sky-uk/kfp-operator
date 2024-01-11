//go:build unit

package vai

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
)

func randomBasicRunIntent() RunIntent {
	return RunIntent{
		PipelineName:    common.RandomNamespacedName(),
		PipelineVersion: common.RandomString(),
	}
}

var _ = Context("VAI Provider", func() {
	Describe("runLabelsFrom", func() {
		It("generates run labels without RunConfigurationName or RunName", func() {
			input := randomBasicRunIntent()

			runLabels := runLabelsFrom(input)

			expectedRunLabels := map[string]string{
				labels.PipelineName:      input.PipelineName.Name,
				labels.PipelineNamespace: input.PipelineName.Namespace,
				labels.PipelineVersion:   input.PipelineVersion,
			}

			Expect(runLabels).To(Equal(expectedRunLabels))
		})

		It("generates run labels with RunConfigurationName", func() {
			input := randomBasicRunIntent()
			input.RunConfigurationName = common.RandomNamespacedName()

			runLabels := runLabelsFrom(input)

			Expect(runLabels[labels.RunConfigurationName]).To(Equal(input.RunConfigurationName.Name))
			Expect(runLabels[labels.RunConfigurationNamespace]).To(Equal(input.RunConfigurationName.Namespace))
			Expect(runLabels).NotTo(HaveKey(labels.RunName))
			Expect(runLabels).NotTo(HaveKey(labels.RunNamespace))
		})

		It("generates run labels with RunName", func() {
			input := randomBasicRunIntent()
			input.RunName = common.RandomNamespacedName()

			runLabels := runLabelsFrom(input)

			Expect(runLabels[labels.RunName]).To(Equal(input.RunName.Name))
			Expect(runLabels[labels.RunNamespace]).To(Equal(input.RunName.Namespace))
			Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationName))
			Expect(runLabels).NotTo(HaveKey(labels.RunConfigurationNamespace))
		})

		It("generates run labels with RunConfigurationName and RunName", func() {
			input := randomBasicRunIntent()
			input.RunConfigurationName = common.RandomNamespacedName()
			input.RunName = common.RandomNamespacedName()

			runLabels := runLabelsFrom(input)

			Expect(runLabels[labels.RunConfigurationName]).To(Equal(input.RunConfigurationName.Name))
			Expect(runLabels[labels.RunConfigurationNamespace]).To(Equal(input.RunConfigurationName.Namespace))
			Expect(runLabels[labels.RunName]).To(Equal(input.RunName.Name))
			Expect(runLabels[labels.RunNamespace]).To(Equal(input.RunName.Namespace))
		})

		It("replaces fullstops with dashes in pipelineVersion", func() {
			input := randomBasicRunIntent()
			input.PipelineVersion = "0.4.0"

			runLabels := runLabelsFrom(input)

			Expect(runLabels[labels.PipelineVersion]).To(Equal("0-4-0"))
		})
	})
})
