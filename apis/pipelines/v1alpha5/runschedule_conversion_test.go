//go:build unit

package v1alpha5

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("RunSchedule Conversion", PropertyBased, func() {
	DefaultProvider = "default-provider"
	DefaultProviderNamespace = "default-provider-namespace"

	var _ = Describe("Roundtrip forward", func() {
		When("status provider is empty", func() {
			Specify("converts to and from the same object", func() {
				noProvider := ""
				src := RandomRunSchedule(noProvider)
				intermediate := &hub.RunSchedule{}
				dst := &RunSchedule{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Name).To(Equal(DefaultProvider))
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Name).To(BeEmpty())
				Expect(intermediate.Status.Provider.Name.Namespace).To(BeEmpty())
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					RunScheduleConversionRemainder{}.ConversionAnnotation(),
				)
				delete(dst.GetAnnotations(), ResourceAnnotations.Provider)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
			})
		})
		When("status provider is non-empty", func() {
			Specify("converts to and from the same object", func() {
				provider := common.RandomString()
				src := RandomRunSchedule(provider)
				intermediate := &hub.RunSchedule{}
				dst := &RunSchedule{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Name).To(Equal(DefaultProvider))
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Name).To(Equal(provider))
				Expect(intermediate.Status.Provider.Name.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					RunScheduleConversionRemainder{}.ConversionAnnotation(),
				)
				delete(dst.GetAnnotations(), ResourceAnnotations.Provider)
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
