//go:build unit
// +build unit

package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Context("RunConfiguration", func() {
	var _ = Describe("ComputeHash", func() {

		Specify("Pipeline should change the hash", func() {
			rcs := RunConfiguration{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.Pipeline = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("ExperimentName should change the hash", func() {
			rcs := RunConfiguration{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.ExperimentName = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("Schedule should change the hash", func() {
			rcs := RunConfiguration{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.Schedule = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("lastKnownPipelineVersion should change the hash", func() {
			rcs := RunConfiguration{}
			hash1 := rcs.ComputeHash()

			rcs.Status.ObservedPipelineVersion = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All RuntimeParameters keys should change the hash", func() {
			rcs := RunConfiguration{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.RuntimeParameters = map[string]string{
				"a": "",
			}
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			rcs.Spec.RuntimeParameters = map[string]string{
				"b": "notempty",
			}
			hash3 := rcs.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})
	})

	var _ = Describe("ComputeVersion", func() {

		Specify("Should have the spec hash only", func() {
			Expect(RunConfiguration{}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})

	var _ = Describe("ExtractPipelineNameVersion", func() {

		Specify("Should return name and version if both exist", func() {
			rcs := RunConfiguration{}
			rcs.Spec.Pipeline = "dummy-pipeline:12345-abcde"
			name, version := rcs.ExtractPipelineNameVersion()

			Expect(name).To(Equal("dummy-pipeline"))
			Expect(version).To(Equal("12345-abcde"))
		})

		Specify("Should return pipeline name and empty version if version doesn't exist", func() {
			rcs := RunConfiguration{}
			rcs.Spec.Pipeline = "my-pipeline"
			name, version := rcs.ExtractPipelineNameVersion()

			Expect(name).To(Equal("my-pipeline"))
			Expect(version).To(Equal(""))
		})
	})
})
