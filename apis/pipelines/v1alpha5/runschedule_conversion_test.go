//go:build unit

package v1alpha5

import (
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("RunSchedule Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip forward", func() {
		Specify("converts string Schedules into Schedule structs", func() {
			// convert from v1alpha5, to v1alpha6, and back
			src := RandomRunSchedule()
			src.Spec.Schedule = "1 1 1 1 1"
			intermediate := &hub.RunSchedule{}
			dst := &RunSchedule{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(intermediate.Spec.Schedule).To(Equal(hub.Schedule{CronExpression: "1 1 1 1 1"}))
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunSchedule()
			now := metav1.Time{Time: time.Now().Truncate(time.Second)}
			src.Spec.Schedule = hub.Schedule{
				CronExpression: "1 1 1 1 1",
				StartTime:      now,
				EndTime:        now,
			}
			intermediate := &RunSchedule{}
			dst := &hub.RunSchedule{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.Spec.Schedule).To(Equal("1 1 1 1 1"))
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
