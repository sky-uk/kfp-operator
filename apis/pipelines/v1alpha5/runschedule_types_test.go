//go:build unit

package v1alpha5

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
)

var _ = Context("RunSchedule", func() {
	var _ = Describe("ComputeHash", func() {

		Specify("Pipeline should change the hash", func() {
			rcs := RunSchedule{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.Pipeline = PipelineIdentifier{Name: "notempty"}
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("ExperimentName should change the hash", func() {
			rcs := RunSchedule{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.ExperimentName = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("Schedule should change the hash", func() {
			rcs := RunSchedule{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.Schedule = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All RuntimeParameters keys should change the hash", func() {
			rcs := RunSchedule{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.RuntimeParameters = []apis.NamedValue{
				{Name: "a", Value: ""},
			}
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			rcs.Spec.RuntimeParameters = []apis.NamedValue{
				{Name: "b", Value: "notempty"},
			}
			hash3 := rcs.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})

		Specify("The original object should not change", PropertyBased, func() {
			rcs := RandomRunSchedule()
			expected := rcs.DeepCopy()
			rcs.ComputeHash()

			Expect(rcs).To(Equal(expected))
		})
	})

	var _ = Describe("ComputeVersion", func() {

		Specify("Should have the spec hash only", func() {
			Expect(RunSchedule{}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})
})
