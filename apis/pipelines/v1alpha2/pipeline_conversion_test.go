package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
)

var _ = Context("Pipeline Conversion", func() {
	var _ = Describe("ConvertTo", func() {
		Specify("Converts Env to a list of NamedValue", func() {
			src := Pipeline{Spec: PipelineSpec{Env: map[string]string{"a": "b", "c": "d"}}}
			expected := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{Env: []v1alpha3.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}

			dst := v1alpha3.Pipeline{}
			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst).To(Equal(expected))
		})

		Specify("Converts BeamArgs to a list of NamedValue", func() {
			src := Pipeline{Spec: PipelineSpec{BeamArgs: map[string]string{"a": "b", "c": "d"}}}
			expected := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{BeamArgs: []v1alpha3.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}

			dst := v1alpha3.Pipeline{}
			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst).To(Equal(expected))
		})

		Specify("Copies all other fields", func() {
			src := Pipeline{
				Spec: PipelineSpec{Image: "image", TfxComponents: "components"},
			}

			dst := v1alpha3.Pipeline{}
			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst.Spec.Image).To(Equal(src.Spec.Image))
			Expect(dst.Spec.TfxComponents).To(Equal(src.Spec.TfxComponents))
		})
	})

	var _ = Describe("ConvertFrom", func() {
		Specify("Converts Env to a map", func() {
			src := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{Env: []v1alpha3.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}
			expected := Pipeline{Spec: PipelineSpec{Env: map[string]string{"a": "b", "c": "d"}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst).To(Equal(expected))
		})

		Specify("Errors when Env contains a duplicate NamedValue", func() {
			src := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{Env: []v1alpha3.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "a", Value: "d"},
			}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).NotTo(Succeed())
		})

		Specify("Converts BeamArgs to a map", func() {
			src := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{BeamArgs: []v1alpha3.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}
			expected := Pipeline{Spec: PipelineSpec{BeamArgs: map[string]string{"a": "b", "c": "d"}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst).To(Equal(expected))
		})

		Specify("Errors when BeamArgs contains a duplicate NamedValue", func() {
			src := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{BeamArgs: []v1alpha3.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "a", Value: "d"},
			}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).NotTo(Succeed())
		})

		Specify("Copies all other fields", func() {
			src := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{Image: "image", TfxComponents: "components"}}
			expected := Pipeline{Spec: PipelineSpec{Image: "image", TfxComponents: "components"}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst).To(Equal(expected))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Does not change between versions", func() {
			src := Pipeline{
				Spec: PipelineSpec{Image: "image", TfxComponents: "components", Env: map[string]string{"a": "b", "c": "d"}, BeamArgs: map[string]string{"a": "b", "c": "d"}},
			}

			dst := v1alpha3.Pipeline{}
			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(src.Spec.ComputeVersion()).To(Equal(dst.Spec.ComputeVersion()))
		})
	})
})
