//go:build unit
// +build unit

package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
)

var _ = Context("Pipeline Conversion", func() {
	var _ = Describe("ConvertTo", func() {
		Specify("Converts Env to a list of NamedValue", func() {
			src := Pipeline{Spec: PipelineSpec{Env: map[string]string{"a": "b", "c": "d"}}}
			dst := v1alpha3.Pipeline{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst.Spec.Env).To(Equal([]apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}))
		})

		Specify("Converts BeamArgs to a list of NamedValue", func() {
			src := Pipeline{Spec: PipelineSpec{BeamArgs: map[string]string{"a": "b", "c": "d"}}}
			dst := v1alpha3.Pipeline{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst.Spec.BeamArgs).To(Equal([]apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}))
		})
	})

	var _ = Describe("ConvertFrom", func() {
		Specify("Converts Env to a map", func() {
			src := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{Env: []apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.Spec.Env).To(Equal(map[string]string{"a": "b", "c": "d"}))
		})

		Specify("Errors when Env contains a duplicate NamedValue", func() {
			src := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{Env: []apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "a", Value: "d"},
			}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).NotTo(Succeed())
		})

		Specify("Converts BeamArgs to a map", func() {
			src := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{BeamArgs: []apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.Spec.BeamArgs).To(Equal(map[string]string{"a": "b", "c": "d"}))
		})

		Specify("Errors when BeamArgs contains a duplicate NamedValue", func() {
			src := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{BeamArgs: []apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "a", Value: "d"},
			}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).NotTo(Succeed())
		})

	})

	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := RandomPipeline()
			intermediate := v1alpha3.Pipeline{}
			dst := Pipeline{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())

			Expect(&dst).To(Equal(src))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Does not change between versions", func() {
			src := RandomPipeline()
			dst := v1alpha3.Pipeline{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(src.Spec.ComputeVersion()).To(Equal(dst.Spec.ComputeVersion()))
		})
	})
})
