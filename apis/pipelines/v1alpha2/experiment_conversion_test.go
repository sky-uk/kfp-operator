//go:build unit
// +build unit

package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
)

var _ = Context("Experiment Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := RandomExperiment()
			intermediate := v1alpha3.Experiment{}
			dst := Experiment{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())

			Expect(&dst).To(Equal(src))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Does not change between versions", func() {
			src := RandomExperiment()
			dst := v1alpha3.Experiment{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(src.Spec.ComputeVersion()).To(Equal(dst.Spec.ComputeVersion()))
		})
	})
})
