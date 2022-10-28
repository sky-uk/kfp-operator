//go:build unit
// +build unit

package v1alpha3

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
)

var _ = Context("Pipeline Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := RandomPipeline()
			intermediate := v1alpha4.Pipeline{}
			dst := Pipeline{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())

			Expect(&dst).To(Equal(src))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Does not change between versions", func() {
			src := RandomPipeline()
			dst := v1alpha4.Pipeline{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(src.ComputeVersion()).To(Equal(dst.ComputeVersion()))
		})
	})
})
