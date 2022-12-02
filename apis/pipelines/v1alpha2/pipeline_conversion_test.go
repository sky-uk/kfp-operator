//go:build unit
// +build unit

package v1alpha2

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("Pipeline Conversion", func() {
	var _ = Describe("ConvertTo", func() {
		Specify("Converts Env to a list of NamedValue", func() {
			src := Pipeline{Spec: PipelineSpec{Env: map[string]string{"a": "b", "c": "d"}}}
			dst := hub.Pipeline{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst.Spec.Env).To(Equal([]apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}))
		})

		Specify("Converts BeamArgs to a list of NamedValue", func() {
			src := Pipeline{Spec: PipelineSpec{BeamArgs: map[string]string{"a": "b", "c": "d"}}}
			dst := hub.Pipeline{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst.Spec.BeamArgs).To(Equal([]apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}))
		})
	})

	var _ = Describe("ConvertFrom", func() {
		Specify("Converts Env to a map", func() {
			src := hub.Pipeline{Spec: hub.PipelineSpec{Env: []apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.Spec.Env).To(Equal(map[string]string{"a": "b", "c": "d"}))
		})

		Specify("Converts BeamArgs to a map", func() {
			src := hub.Pipeline{Spec: hub.PipelineSpec{BeamArgs: []apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.Spec.BeamArgs).To(Equal(map[string]string{"a": "b", "c": "d"}))
		})

		Specify("Removes duplicates and adds remainder annotation when BeamArgs contains a duplicate NamedValue", func() {
			src := hub.Pipeline{Spec: hub.PipelineSpec{BeamArgs: []apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "a", Value: "c"},
			}, Env: []apis.NamedValue{
				{Name: "d", Value: "e"},
				{Name: "d", Value: "f"},
			}}}
			dst := Pipeline{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.Spec.BeamArgs).To(Equal(map[string]string{"a": "b"}))
			Expect(dst.Spec.Env).To(Equal(map[string]string{"d": "e"}))
		})
	})

	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomPipeline()
			intermediate := &Pipeline{}
			dst := &hub.Pipeline{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(src.Spec.BeamArgs).To(ConsistOf(dst.Spec.BeamArgs))
			src.Spec.BeamArgs = nil
			dst.Spec.BeamArgs = nil

			Expect(src.Spec.Env).To(ConsistOf(dst.Spec.Env))
			src.Spec.Env = nil
			dst.Spec.Env = nil

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})

		Specify("Duplicate entries are preserved on the roundtrip", func() {
			src := hub.Pipeline{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: hub.PipelineSpec{
					BeamArgs: []apis.NamedValue{
						{Name: "a", Value: "b"},
						{Name: "a", Value: "c"},
					}, Env: []apis.NamedValue{
						{Name: "d", Value: "e"},
						{Name: "d", Value: "f"},
					},
				},
			}

			intermediate := Pipeline{}
			dst := hub.Pipeline{}

			Expect(intermediate.ConvertFrom(&src)).To(Succeed())
			Expect(intermediate.ConvertTo(&dst)).To(Succeed())

			Expect(src).To(Equal(dst))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Does not change between versions", func() {
			src := hub.RandomPipeline()
			dst := Pipeline{}

			Expect(dst.ConvertFrom(src)).To(Succeed())
			Expect(src.ComputeVersion()).To(Equal(dst.ComputeVersion()))
		})
	})
})
