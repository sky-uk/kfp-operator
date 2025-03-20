//go:build unit

package v1alpha5

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("Experiment Conversion", PropertyBased, func() {
	DefaultProvider = "default-provider"
	DefaultProviderNamespace = "default-provider-namespace"

	var _ = Describe("Roundtrip forward", func() {
		When("status provider is empty", func() {
			It("converts to and from the same object", func() {
				noProvider := ""
				src := RandomExperiment(noProvider)
				intermediate := &hub.Experiment{}
				dst := &Experiment{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Name).To(Equal(DefaultProvider))
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Name).To(BeEmpty())
				Expect(intermediate.Status.Provider.Name.Namespace).To(BeEmpty())
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					ExperimentConversionRemainder{}.ConversionAnnotation(),
				)
				delete(dst.GetAnnotations(), ResourceAnnotations.Provider)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
			})
		})
		When("status provider is non-empty", func() {
			It("converts to and from the same object", func() {
				provider := common.RandomString()
				src := RandomExperiment(provider)
				intermediate := &hub.Experiment{}
				dst := &Experiment{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Name).To(Equal(DefaultProvider))
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Name).To(Equal(provider))
				Expect(intermediate.Status.Provider.Name.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				Expect(getProviderAnnotation(dst)).To(Equal(DefaultProvider))
				delete(
					dst.GetAnnotations(),
					ExperimentConversionRemainder{}.ConversionAnnotation(),
				)
				delete(dst.GetAnnotations(), ResourceAnnotations.Provider)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
			})
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomExperiment(common.RandomNamespacedName())
			intermediate := &Experiment{}
			dst := &hub.Experiment{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
