package v1

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
)

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Api Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = Describe("Pipeline Version", func() {
	When("Specifying values for fields", func() {
		It("Should change the version", func() {
			pipelineSpec := PipelineSpec{}
			version1, err := ComputeVersion(pipelineSpec)
			Expect(err).NotTo(HaveOccurred())
			pipelineSpec.Image = "image"
			version2, err := ComputeVersion(pipelineSpec)
			Expect(err).NotTo(HaveOccurred())
			pipelineSpec.TfxComponents = "components"
			version3, err := ComputeVersion(pipelineSpec)
			Expect(err).NotTo(HaveOccurred())
			pipelineSpec.Env = map[string]string{
				"aKey": "aValue",
			}
			version4, err := ComputeVersion(pipelineSpec)
			Expect(err).NotTo(HaveOccurred())
			pipelineSpec.Env["bKey"] = ""
			version5, err := ComputeVersion(pipelineSpec)
			Expect(err).NotTo(HaveOccurred())

			Expect(version1).NotTo(Equal(version2))
			Expect(version2).NotTo(Equal(version3))
			Expect(version3).NotTo(Equal(version4))
			Expect(version4).NotTo(Equal(version5))
		})
	})

	When("Not specifying values for Env", func() {
		It("Should not change the version", func() {
			pipelineSpec := PipelineSpec{}
			version1, err := ComputeVersion(pipelineSpec)
			Expect(err).NotTo(HaveOccurred())
			pipelineSpec.Env = make(map[string]string)
			version2, err := ComputeVersion(pipelineSpec)
			Expect(err).NotTo(HaveOccurred())

			Expect(version1).To(Equal(version2))
		})
	})
})
