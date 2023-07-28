//go:build unit

package v1alpha4

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
)

var _ = Context("RunConfiguration Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip forward", func() {
		Specify("converts empty Schedule into no triggers", func() {
			src := RandomRunConfiguration()
			src.Spec.Schedule = ""
			src.Status.ProviderId.Id = ""
			src.Status.Version = ""
			intermediate := &hub.RunConfiguration{}
			dst := &RunConfiguration{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(intermediate.Spec.Triggers.Schedules).To(BeEmpty())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunConfiguration()
			hub.WithValueFrom(&src.Spec.Run)
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst.Spec.Run.RuntimeParameters).To(ConsistOf(src.Spec.Run.RuntimeParameters))
			dst.Spec.Run.RuntimeParameters = nil
			src.Spec.Run.RuntimeParameters = nil
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
