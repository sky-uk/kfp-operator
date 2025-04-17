//go:build unit

package v1alpha5

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Context("Provider Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomProvider()

			intermediate := &Provider{}
			dst := &hub.Provider{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Conversion failure", func() {
		Specify("ConvertTo fails when there is no serviceImage annotation", func() {
			src := RandomProvider()
			dst := &hub.Provider{}

			Expect(src.ConvertTo(dst)).To(Not(Succeed()))
		})
	})
})
