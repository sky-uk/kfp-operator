//go:build unit
// +build unit

package v1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ComputeVersion", func() {

	Specify("Image should change the version", func() {
		pipelineSpec := PipelineSpec{}
		version1 := ComputeVersion(pipelineSpec)

		pipelineSpec.Image = "notempty"
		version2 := ComputeVersion(pipelineSpec)

		Expect(version1).NotTo(Equal(version2))
	})

	Specify("TfxComponents should change the version", func() {
		pipelineSpec := PipelineSpec{}
		version1 := ComputeVersion(pipelineSpec)

		pipelineSpec.TfxComponents = "notempty"
		version2 := ComputeVersion(pipelineSpec)

		Expect(version1).NotTo(Equal(version2))
	})

	Specify("All Env keys should change the version", func() {
		pipelineSpec := PipelineSpec{}
		version1 := ComputeVersion(pipelineSpec)

		pipelineSpec.Env = map[string]string{
			"a": "",
		}
		version2 := ComputeVersion(pipelineSpec)

		pipelineSpec.Env = map[string]string{
			"b": "notempty",
		}
		version3 := ComputeVersion(pipelineSpec)

		Expect(version1).NotTo(Equal(version2))
		Expect(version2).NotTo(Equal(version3))
	})
})
