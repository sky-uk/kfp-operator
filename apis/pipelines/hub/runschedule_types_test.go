//go:build unit

package v1beta1

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
			rs := RunSchedule{}
			hash1 := rs.ComputeHash()

			rs.Spec.Pipeline = PipelineIdentifier{Name: "notempty"}
			hash2 := rs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("ExperimentName should change the hash", func() {
			rs := RunSchedule{}
			hash1 := rs.ComputeHash()

			rs.Spec.ExperimentName = "notempty"
			hash2 := rs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("Schedule CronExpression should change the hash", func() {
			rs := RunSchedule{}
			hash1 := rs.ComputeHash()

			rs.Spec.Schedule.CronExpression = "notempty"
			hash2 := rs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("Schedule StartTime should change the hash", func() {
			rs := RunSchedule{}
			hash1 := rs.ComputeHash()

			startTime := metav1.NewTime(time.Now())
			rs.Spec.Schedule.StartTime = &startTime
			hash2 := rs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("Schedule EndTime should change the hash", func() {
			rs := RunSchedule{}
			hash1 := rs.ComputeHash()

			endTime := metav1.NewTime(time.Now())
			rs.Spec.Schedule.EndTime = &endTime
			hash2 := rs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All RuntimeParameters keys should change the hash", func() {
			rs := RunSchedule{}
			hash1 := rs.ComputeHash()

			rs.Spec.RuntimeParameters = []apis.NamedValue{
				{Name: "a", Value: ""},
			}
			hash2 := rs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			rs.Spec.RuntimeParameters = []apis.NamedValue{
				{Name: "b", Value: "notempty"},
			}
			hash3 := rs.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})

		Specify("The original object should not change", PropertyBased, func() {
			rs := RandomRunSchedule(apis.RandomLowercaseString())
			expected := rs.DeepCopy()
			rs.ComputeHash()

			Expect(rs).To(Equal(expected))
		})
	})

	var _ = Describe("ComputeVersion", func() {

		Specify("Should have the spec hash only", func() {
			Expect(RunSchedule{}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})
})
