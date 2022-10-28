//go:build unit
// +build unit

package v1alpha3

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
)

var _ = Context("RunConfiguration Conversion", func() {
	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := RandomRunConfiguration()
			intermediate := hub.RunConfiguration{}
			dst := RunConfiguration{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())

			Expect(&dst).To(Equal(src))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Does not change between versions", func() {
			src := RunConfiguration{
				Status: RunConfigurationStatus{Status: Status{KfpId: apis.RandomString()}},
			}
			dst := hub.RunConfiguration{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(src.ComputeVersion()).To(Equal(dst.ComputeVersion()))
		})
	})
})
