//go:build unit

package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("RunConfiguration", func() {
	var _ = Describe("ComputeHash", func() {

		Specify("Pipeline should change the hash", func() {
			rc := RunConfiguration{}
			hash1 := rc.ComputeHash()

			rc.Spec.Pipeline = PipelineIdentifier{Name: "notempty"}
			hash2 := rc.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("ExperimentName should change the hash", func() {
			rc := RunConfiguration{}
			hash1 := rc.ComputeHash()

			rc.Spec.ExperimentName = "notempty"
			hash2 := rc.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("Schedule should change the hash", func() {
			rc := RunConfiguration{}
			hash1 := rc.ComputeHash()

			rc.Spec.Schedule = "notempty"
			hash2 := rc.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("ObservedPipelineVersion should change the hash", func() {
			rc := RunConfiguration{}
			hash1 := rc.ComputeHash()

			rc.Status.ObservedPipelineVersion = "notempty"
			hash2 := rc.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All RuntimeParameters keys should change the hash", func() {
			rc := RunConfiguration{}
			hash1 := rc.ComputeHash()

			rc.Spec.RuntimeParameters = map[string]string{
				"a": "",
			}
			hash2 := rc.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			rc.Spec.RuntimeParameters = map[string]string{
				"b": "notempty",
			}
			hash3 := rc.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})
	})

	var _ = Describe("ComputeVersion", func() {

		Specify("Should have the spec hash only", func() {
			Expect(RunConfiguration{}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})
})
