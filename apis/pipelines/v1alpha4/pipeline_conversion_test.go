//go:build unit
// +build unit

package v1alpha4

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
)

var _ = Context("Pipeline Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomPipeline()
			intermediate := &Pipeline{}
			dst := &hub.Pipeline{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
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
