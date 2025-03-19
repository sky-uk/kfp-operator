//go:build unit

package v1alpha5

import (
	"fmt"
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
		Specify("converts to and from the same object", func() {
			src := RandomRunConfiguration()
			src.Status.Provider = DefaultProvider

			intermediate := &hub.RunConfiguration{}
			dst := &RunConfiguration{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(intermediate.Spec.Run.Provider.Name).To(Equal(DefaultProvider))
			Expect(intermediate.Spec.Run.Provider.Namespace).To(Equal(DefaultProviderNamespace))
			Expect(intermediate.Status.Provider.Name).To(Equal(DefaultProvider))
			Expect(intermediate.Status.Provider.Namespace).To(Equal(DefaultProviderNamespace))
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())
			delete(
				dst.GetAnnotations(),
				RunConfigurationConversionRemainder{}.ConversionAnnotation(),
			)
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunConfiguration(common.RandomNamespacedName())
			hub.WithValueFrom(&src.Spec.Run)
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			fmt.Printf("intermediate.Spec.Provider.Name: %v\n", intermediate.Annotations)
			Expect(intermediate.ConvertTo(dst)).To(Succeed())
			fmt.Printf("dst.Status.Provider.Name: %v\n", dst.Status.Provider.Name)
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
