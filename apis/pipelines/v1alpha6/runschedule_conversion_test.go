//go:build unit

package v1alpha6

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("RunSchedule Conversion", PropertyBased, func() {
	DefaultProviderNamespace = "default-provider-namespace"

	var _ = Describe("Roundtrip forward", func() {
		When("status provider is empty", func() {
			It("converts to and from the same object", func() {
				src := RandomRunSchedule(apis.RandomLowercaseString())
				src.Status.Provider.Name = ""
				intermediate := &hub.RunSchedule{}
				dst := &RunSchedule{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Namespace).To(BeEmpty())
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					RunScheduleConversionRemainder{}.ConversionAnnotation(),
				)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
			})
		})
		When("status provider is non-empty", func() {
			It("converts to and from the same object", func() {
				src := RandomRunSchedule(apis.RandomLowercaseString())
				intermediate := &hub.RunSchedule{}
				dst := &RunSchedule{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					RunScheduleConversionRemainder{}.ConversionAnnotation(),
				)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
			})
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunSchedule(common.RandomNamespacedName())
			intermediate := &RunSchedule{}
			dst := &hub.RunSchedule{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
