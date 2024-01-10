//go:build unit

package vai

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
)

func expectedDefaultLabels(runIntent RunIntent) map[string]string {
	return map[string]string{
		labels.PipelineName:      runIntent.PipelineName.Name,
		labels.PipelineNamespace: runIntent.PipelineName.Namespace,
		labels.PipelineVersion:   runIntent.PipelineVersion,
	}
}

func randomBasicRunIntent() RunIntent {
	return RunIntent{
		PipelineName:      common.RandomNamespacedName(),
		PipelineVersion:   common.RandomString(),
		RuntimeParameters: nil,
		Artifacts:         nil,
	}
}

var _ = Context("VAI Provider", func() {
	Describe("runLabelsFrom", func() {

		It("generates runLabel without RunConfigurationName or RunName", func() {
			input := randomBasicRunIntent()
			runLabels := runLabelsFrom(input)
			Expect(runLabels).To(Equal(expectedDefaultLabels(input)))
		})

		It("generates runLabel with RunConfigurationName", func() {
			input := randomBasicRunIntent()
			input.RunConfigurationName = common.RandomNamespacedName()

			expected := expectedDefaultLabels(input)
			expected[labels.RunConfigurationName] = input.RunConfigurationName.Name
			expected[labels.RunConfigurationNamespace] = input.RunConfigurationName.Namespace

			runLabels := runLabelsFrom(input)
			Expect(runLabels).To(Equal(expected))
		})

		It("generates runLabel with RunName", func() {
			input := randomBasicRunIntent()
			input.RunName = common.RandomNamespacedName()
			expected := expectedDefaultLabels(input)
			expected[labels.RunName] = input.RunName.Name
			expected[labels.RunNamespace] = input.RunName.Namespace

			runLabels := runLabelsFrom(input)
			Expect(runLabels).To(Equal(expected))
		})

		// lowercase letters, numbers, dashes and underscores
		It("replaces fullstops with dashes in pipelineVersion", func() {
			input := randomBasicRunIntent()
			input.PipelineVersion = "0.4.0"
			expected := expectedDefaultLabels(input)
			expected[labels.PipelineVersion] = "0-4-0"

			runLabels := runLabelsFrom(input)
			Expect(runLabels).To(Equal(expected))
		})
	})
})
