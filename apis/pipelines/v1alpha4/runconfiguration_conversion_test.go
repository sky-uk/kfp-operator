//go:build unit
// +build unit

package v1alpha4

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
)

var _ = Context("RunConfiguration Conversion", func() {
	var _ = Describe("Roundtrip forward", func() {
		Specify("converts to and from the same object", func() {
			src := RandomRunConfiguration()
			src.Status.ProviderId.Id = ""
			src.Status.Version = ""
			intermediate := &hub.RunConfiguration{}
			dst := &RunConfiguration{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunConfiguration()
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})

		//TODO: decide whether to preserve order
		Specify("adds the first Schedule trigger to the front", func() {
			src := hub.RandomRunConfiguration()
			onChange := hub.RandomOnChangeTrigger()
			schedule1 := hub.RandomCronTrigger()
			schedule2 := hub.RandomCronTrigger()
			src.Spec.Triggers = []hub.Trigger{onChange, schedule1, schedule2}
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst.Spec.Triggers).To(Equal([]hub.Trigger{schedule1, onChange, schedule2}))
		})

		Specify("preserves excess schedule triggers in the conversion annotation", func() {
			src := hub.RandomRunConfiguration()
			src.Spec.Triggers = []hub.Trigger{hub.RandomCronTrigger(), hub.RandomCronTrigger()}
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})

		Specify("preserves all onChange triggers in the conversion annotation", func() {
			src := hub.RandomRunConfiguration()
			src.Spec.Triggers = []hub.Trigger{hub.RandomOnChangeTrigger(), hub.RandomOnChangeTrigger()}
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
