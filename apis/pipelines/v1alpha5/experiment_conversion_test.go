//go:build unit

package v1alpha5

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Context("Experiment Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip forward", func() {
		Specify("converts to and from the same object using default provider", func() {
			src := RandomExperiment()
			DefaultProvider = "default-provider"
			intermediate := &hub.Experiment{}
			dst := &Experiment{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())
			Expect(getProviderAnnotation(dst)).To(Equal(DefaultProvider))
		})

		Specify("converts to and from the same object", func() {
			src := RandomExperiment()
			setProviderAnnotation(apis.RandomLowercaseString(), &src.ObjectMeta)
			intermediate := &hub.Experiment{}
			dst := &Experiment{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomExperiment(apis.RandomLowercaseString())
			intermediate := &Experiment{}
			dst := &hub.Experiment{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
