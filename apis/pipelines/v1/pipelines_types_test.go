//go:build unit
// +build unit

package v1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ComputeHash", func() {

	Specify("Image should change the hash", func() {
		pipelineSpec := PipelineSpec{}
		hash1 := pipelineSpec.ComputeHash()

		pipelineSpec.Image = "notempty"
		hash2 := pipelineSpec.ComputeHash()

		Expect(hash1).NotTo(Equal(hash2))
	})

	Specify("TfxComponents should change the hash", func() {
		pipelineSpec := PipelineSpec{}
		hash1 := pipelineSpec.ComputeHash()

		pipelineSpec.TfxComponents = "notempty"
		hash2 := pipelineSpec.ComputeHash()

		Expect(hash1).NotTo(Equal(hash2))
	})

	Specify("All Env keys should change the hash", func() {
		pipelineSpec := PipelineSpec{}
		hash1 := pipelineSpec.ComputeHash()

		pipelineSpec.Env = map[string]string{
			"a": "",
		}
		hash2 := pipelineSpec.ComputeHash()

		Expect(hash1).NotTo(Equal(hash2))

		pipelineSpec.Env = map[string]string{
			"b": "notempty",
		}
		hash3 := pipelineSpec.ComputeHash()

		Expect(hash2).NotTo(Equal(hash3))
	})

	Specify("All BeamArgs keys should change the hash", func() {
		pipelineSpec := PipelineSpec{}
		hash1 := pipelineSpec.ComputeHash()

		pipelineSpec.BeamArgs = map[string]string{
			"a": "",
		}
		hash2 := pipelineSpec.ComputeHash()

		pipelineSpec.BeamArgs = map[string]string{
			"b": "notempty",
		}
		hash3 := pipelineSpec.ComputeHash()

		Expect(hash1).NotTo(Equal(hash2))
		Expect(hash2).NotTo(Equal(hash3))
	})
})

var _ = Describe("ComputeVersion", func() {

	Specify("Should start with image tag if present", func() {
		pipelineSpec := PipelineSpec{
			Image: "image:42",
		}

		version := pipelineSpec.ComputeVersion()

		Expect(version).To(MatchRegexp("^42-[a-z0-9]{6}$"))
	})

	Specify("Should strip the image tag when domain and path present", func() {
		pipelineSpec := PipelineSpec{
			Image: "docker.io/baz/bar/image:baz",
		}

		version := pipelineSpec.ComputeVersion()

		Expect(version).To(MatchRegexp("^baz-[a-z0-9]{6}$"))
	})

	Specify("Untagged images should have the spec hash only", func() {
		pipelineSpec := PipelineSpec{
			Image: "image",
		}

		version := pipelineSpec.ComputeVersion()

		Expect(version).To(MatchRegexp("^[a-z0-9]{6}$"))
	})

	Specify("Malformed image names should have the spec hash only", func() {
		pipelineSpec := PipelineSpec{
			Image: ":",
		}

		version := pipelineSpec.ComputeVersion()

		Expect(version).To(MatchRegexp("^[a-z0-9]{6}$"))
	})
})
