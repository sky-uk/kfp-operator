//go:build unit

package v1alpha6

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/pkg/common"
)

var _ = Context("Experiment Conversion", PropertyBased, func() {
	DefaultProviderNamespace = "default-provider-namespace"

	var _ = Describe("Roundtrip forward", func() {
		When("status provider is empty", func() {
			It("converts to and from the same object", func() {
				src := RandomExperiment(apis.RandomLowercaseString())
				src.Status.Provider.Name = ""
				intermediate := &hub.Experiment{}
				dst := &Experiment{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Namespace).To(BeEmpty())
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					ExperimentConversionRemainder{}.ConversionAnnotation(),
				)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty(), syncStateComparer))
			})
		})

		When("status provider is non-empty", func() {
			It("converts to and from the same object", func() {
				src := RandomExperiment(apis.RandomLowercaseString())
				intermediate := &hub.Experiment{}
				dst := &Experiment{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					ExperimentConversionRemainder{}.ConversionAnnotation(),
				)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty(), syncStateComparer))
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
