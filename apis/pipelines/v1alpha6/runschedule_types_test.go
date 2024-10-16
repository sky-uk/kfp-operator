//go:build unit

package v1alpha6

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

		Specify("Schedule CronExpression should change the hash", func() {
			rcs := RunSchedule{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.Schedule.CronExpression = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("Schedule StartTime should change the hash", func() {
			rcs := RunSchedule{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.Schedule.StartTime = metav1.NewTime(time.Date(1996, 4, 11, 6, 9, 0, 0, time.UTC))
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("Schedule CronExpression should change the hash", func() {
			rcs := RunSchedule{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.Schedule.EndTime = metav1.NewTime(time.Date(1997, 11, 13, 8, 0, 0, 0, time.UTC))
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
