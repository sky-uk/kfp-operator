//go:build unit

package v1alpha5

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("RunConfiguration Conversion", PropertyBased, func() {
	DefaultProvider = "default-provider"
	DefaultProviderNamespace = "default-provider-namespace"

	var _ = Describe("Roundtrip forward", func() {
		When("status provider is empty", func() {
			Specify("converts to and from the same object", func() {
				provider := common.RandomString()
				noStatusProvider := ""
				src := RandomRunConfiguration(noStatusProvider)
				setProviderAnnotation(provider, &src.ObjectMeta)
				intermediate := &hub.RunConfiguration{}
				dst := &RunConfiguration{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Run.Provider.Name).To(Equal(provider))
				Expect(intermediate.Spec.Run.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name).To(BeEmpty())
				Expect(intermediate.Status.Provider.Namespace).To(BeEmpty())
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					RunConfigurationConversionRemainder{}.ConversionAnnotation(),
				)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
			})
		})
		When("status provider is non-empty", func() {
			Specify("converts to and from the same object", func() {
				provider := common.RandomString()
				src := RandomRunConfiguration(provider)
				setProviderAnnotation(provider, &src.ObjectMeta)
				intermediate := &hub.RunConfiguration{}
				dst := &RunConfiguration{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Run.Provider.Name).To(Equal(provider))
				Expect(intermediate.Spec.Run.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name).To(Equal(provider))
				Expect(intermediate.Status.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					RunConfigurationConversionRemainder{}.ConversionAnnotation(),
				)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
			})
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			provider := common.RandomNamespacedName()
			src := hub.RandomRunConfiguration(provider)
			hub.WithValueFrom(&src.Spec.Run)
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.Status.Provider).To(Equal(provider.Name))
			Expect(intermediate.ConvertTo(dst)).To(Succeed())
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
