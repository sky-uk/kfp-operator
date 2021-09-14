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
			version1 := ComputeVersion(pipelineSpec)

			pipelineSpec.Image = "image"
			version2 := ComputeVersion(pipelineSpec)

			pipelineSpec.TfxComponents = "components"
			version3 := ComputeVersion(pipelineSpec)

			pipelineSpec.Env = map[string]string{
				"aKey": "aValue",
			}
			version4 := ComputeVersion(pipelineSpec)

			pipelineSpec.Env["bKey"] = ""
			version5 := ComputeVersion(pipelineSpec)

			Expect(version1).NotTo(Equal(version2))
			Expect(version2).NotTo(Equal(version3))
			Expect(version3).NotTo(Equal(version4))
			Expect(version4).NotTo(Equal(version5))
		})
	})

	When("Not specifying values for Env", func() {
		It("Should not change the version", func() {
			pipelineSpec := PipelineSpec{}
			version1 := ComputeVersion(pipelineSpec)

			pipelineSpec.Env = make(map[string]string)
			version2 := ComputeVersion(pipelineSpec)

			Expect(version1).To(Equal(version2))
		})
	})
})
