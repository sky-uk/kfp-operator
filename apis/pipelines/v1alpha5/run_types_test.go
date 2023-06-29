//go:build unit
// +build unit

package v1alpha5

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Run", func() {
	var _ = Describe("ComputeHash", func() {

		Specify("Pipeline should change the hash", func() {
			run := Run{}
			hash1 := run.ComputeHash()

			run.Spec.Pipeline = PipelineIdentifier{Name: "notempty"}
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("ExperimentName should change the hash", func() {
			run := Run{}
			hash1 := run.ComputeHash()

			run.Spec.ExperimentName = "notempty"
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("ObservedPipelineVersion should change the hash", func() {
			run := Run{}
			hash1 := run.ComputeHash()

			run.Status.ObservedPipelineVersion = "notempty"
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All RuntimeParameters keys should change the hash", func() {
			run := Run{}
			hash1 := run.ComputeHash()

			run.Spec.RuntimeParameters = []RuntimeParameter{
				{Name: "a", Value: ""},
			}
			hash2 := run.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			run.Spec.RuntimeParameters = []RuntimeParameter{
				{Name: "b", Value: "notempty"},
			}
			hash3 := run.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})

		Specify("The original object should not change", PropertyBased, func() {
			run := RandomRun()
			expected := run.DeepCopy()
			run.ComputeHash()

			Expect(run).To(Equal(expected))
		})
	})

	var _ = Describe("ComputeVersion", func() {

		Specify("Should have the spec hash only", func() {
			Expect(Run{}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})
})
