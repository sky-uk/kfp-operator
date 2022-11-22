//go:build unit
// +build unit

package v1alpha3

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
)

var _ = Context("RunConfiguration Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunConfiguration()
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(dst).To(Equal(src))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Does not change between versions", func() {
			src := hub.RandomRunConfiguration()
			dst := RunConfiguration{}

			Expect(dst.ConvertFrom(src)).To(Succeed())
			Expect(src.ComputeVersion()).To(Equal(dst.ComputeVersion()))
		})
	})
})