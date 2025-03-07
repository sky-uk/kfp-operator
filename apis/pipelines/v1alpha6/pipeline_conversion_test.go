//go:build unit

package v1alpha6

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Context("Pipeline Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip forward", func() {

		Specify("converts to and from the same object", func() {
			src := RandomPipeline(apis.RandomLowercaseString())
			intermediate := &hub.Pipeline{}
			dst := &Pipeline{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomPipeline(apis.RandomLowercaseString())
			intermediate := &Pipeline{}
			dst := &hub.Pipeline{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})

		Specify("converts to and from the same object when the framework is tfx", func() {
			src := hub.RandomPipeline(apis.RandomLowercaseString())
			src.Spec.Framework.Type = "tfx"
			intermediate := &Pipeline{}
			dst := &hub.Pipeline{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Conversion failure", func() {
		Specify("ConvertFrom fails when the framework is tfx and there is no components parameter", func() {
			src := hub.RandomPipeline(apis.RandomLowercaseString())
			src.Spec.Framework.Type = "tfx"
			src.Spec.Framework.Parameters = nil
			intermediate := &Pipeline{}

			Expect(intermediate.ConvertFrom(src)).To(Not(Succeed()))
		})

		Specify("ConvertTo fails when there are no annotations and tfx components isn't set", func() {
			src := RandomPipeline(apis.RandomLowercaseString())
			src.Spec.TfxComponents = ""
			dst := &hub.Pipeline{}

			Expect(src.ConvertTo(dst)).To(Not(Succeed()))
		})
	})
})
