//go:build unit

package v1alpha5

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
)

var _ = Context("RunConfiguration Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip forward", func() {
		Specify("converts string Schedules into Schedule structs", func() {
			// convert from v1alpha5, to v1alpha6, and back
			src := RandomRunConfiguration()
			src.Spec.Triggers.Schedules = []string{"1 1 1 1 1"}
			intermediate := &hub.RunConfiguration{}
			dst := &RunConfiguration{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(intermediate.Spec.Triggers.Schedules).To(Equal([]hub.Schedule{{CronExpression: "1 1 1 1 1"}}))
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunConfiguration()
			src.Spec.Triggers.Schedules = []hub.Schedule{{CronExpression: "1 1 1 1 1"}}
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.Spec.Triggers.Schedules).To(Equal([]string{"1 1 1 1 1"}))
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
