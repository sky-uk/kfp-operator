//go:build unit

package v1alpha5

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"strings"
)

func namedValueSort(a, b apis.NamedValue) bool {
	return strings.Compare(a.Name, b.Name) < 0
}

var _ = Context("Pipeline Conversion", PropertyBased, func() {
	DefaultProvider = "default-provider"
	DefaultProviderNamespace = "default-provider-namespace"

	var _ = Describe("Roundtrip forward", func() {
		When("status provider is empty", func() {
			Specify("converts to and from the same object", func() {
				provider := common.RandomString()
				noStatusProvider := ""
				src := RandomPipeline(noStatusProvider)
				setProviderAnnotation(provider, &src.ObjectMeta)
				intermediate := &hub.Pipeline{}
				dst := &Pipeline{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Name).To(Equal(provider))
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Name).To(BeEmpty())
				Expect(intermediate.Status.Provider.Name.Namespace).To(BeEmpty())
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					PipelineConversionRemainder{}.ConversionAnnotation(),
				)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty(), cmpopts.SortSlices(namedValueSort), syncStateComparer))
			})
		})
		When("status provider is non-empty", func() {
			Specify("converts to and from the same object", func() {
				provider := common.RandomString()
				src := RandomPipeline(provider)
				setProviderAnnotation(provider, &src.ObjectMeta)
				intermediate := &hub.Pipeline{}
				dst := &Pipeline{}

				Expect(src.ConvertTo(intermediate)).To(Succeed())
				Expect(intermediate.Spec.Provider.Name).To(Equal(provider))
				Expect(intermediate.Spec.Provider.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(intermediate.Status.Provider.Name.Name).To(Equal(provider))
				Expect(intermediate.Status.Provider.Name.Namespace).To(Equal(DefaultProviderNamespace))
				Expect(dst.ConvertFrom(intermediate)).To(Succeed())
				delete(
					dst.GetAnnotations(),
					PipelineConversionRemainder{}.ConversionAnnotation(),
				)
				Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty(), cmpopts.SortSlices(namedValueSort), syncStateComparer))
			})
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object when the framework is tfx", func() {
			provider := common.RandomNamespacedName()
			src := hub.RandomPipeline(provider)
			hub.AddTfxValues(&src.Spec)
			intermediate := &Pipeline{}
			dst := &hub.Pipeline{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.Status.ProviderId.Provider).To(Equal(provider.Name))
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty(), cmpopts.SortSlices(namedValueSort)))
		})

		Specify("converts to and from the same object when the framework is not tfx", func() {
			provider := common.RandomNamespacedName()
			src := hub.RandomPipeline(provider)
			src.Spec.Framework.Type = "some-other-framework"
			intermediate := &Pipeline{}
			dst := &hub.Pipeline{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.Status.ProviderId.Provider).To(Equal(provider.Name))
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty(), cmpopts.SortSlices(namedValueSort)))
		})
	})

	var _ = Describe("Conversion failure", func() {
		Specify("ConvertTo fails when there are no annotations and tfx components isn't set", func() {
			src := RandomPipeline(common.RandomString())
			src.Spec.TfxComponents = ""
			dst := &hub.Pipeline{}

			Expect(src.ConvertTo(dst)).To(Not(Succeed()))
		})
	})
})
