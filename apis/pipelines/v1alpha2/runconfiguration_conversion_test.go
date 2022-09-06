//go:build unit
// +build unit

package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
)

var _ = Context("RunConfiguration Conversion", func() {
	var _ = Describe("ConvertTo", func() {

		Specify("Converts RuntimeParameters to a list of NamedValue", func() {
			src := RunConfiguration{Spec: RunConfigurationSpec{RuntimeParameters: map[string]string{"a": "b", "c": "d"}}}
			dst := v1alpha3.RunConfiguration{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst.Spec.RuntimeParameters).To(Equal([]apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}))
		})
	})

	var _ = Describe("ConvertFrom", func() {

		Specify("Converts RuntimeParameters to a map", func() {
			src := v1alpha3.RunConfiguration{Spec: v1alpha3.RunConfigurationSpec{
				RuntimeParameters: []apis.NamedValue{
					{Name: "a", Value: "b"},
					{Name: "c", Value: "d"},
				}}}
			dst := RunConfiguration{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.Spec.RuntimeParameters).To(Equal(map[string]string{"a": "b", "c": "d"}))
		})

		Specify("Errors when RuntimeParameters contains a duplicate NamedValue", func() {
			src := v1alpha3.RunConfiguration{Spec: v1alpha3.RunConfigurationSpec{
				RuntimeParameters: []apis.NamedValue{
					{Name: "a", Value: "b"},
					{Name: "a", Value: "d"},
				}}}
			dst := RunConfiguration{}

			Expect(dst.ConvertFrom(&src)).NotTo(Succeed())
		})
	})

	var _ = Describe("Roundtrip", func() {
		Specify("converts to and from the same object", func() {
			src := RandomRunConfiguration()
			intermediate := v1alpha3.RunConfiguration{}
			dst := RunConfiguration{}

			Expect(src.ConvertTo(&intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(&intermediate)).To(Succeed())

			Expect(&dst).To(Equal(src))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Does not change between versions", func() {
			src := RunConfiguration{
				Spec: RunConfigurationSpec{RuntimeParameters: map[string]string{"a": "b", "c": "d"}},
			}
			dst := v1alpha3.RunConfiguration{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(src.ComputeVersion()).To(Equal(dst.ComputeVersion()))
		})
	})
})
