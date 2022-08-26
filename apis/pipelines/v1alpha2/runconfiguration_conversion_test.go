package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("RunConfiguration Conversion", func() {
	var _ = Describe("ConvertTo", func() {

		Specify("Converts RuntimeParameters to a list of NamedValue", func() {
			src := RunConfiguration{Spec: RunConfigurationSpec{RuntimeParameters: map[string]string{"a": "b", "c": "d"}}}
			expected := v1alpha3.RunConfiguration{Spec: v1alpha3.RunConfigurationSpec{RuntimeParameters: []apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}

			dst := v1alpha3.RunConfiguration{}
			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst).To(Equal(expected))
		})

		Specify("Copies all other fields", func() {
			src := RunConfiguration{
				ObjectMeta: v1.ObjectMeta{Name: "runconfiguration"},
				Status: RunConfigurationStatus{
					Status: apis.Status{
						Version:              "1",
						KfpId:                "id",
						SynchronizationState: apis.Succeeded,
					},
					ObservedPipelineVersion: "1",
				},
			}

			dst := v1alpha3.RunConfiguration{}
			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst.ObjectMeta).To(Equal(src.ObjectMeta))
			Expect(dst.Status.Status).To(Equal(src.Status.Status))
			Expect(dst.Status.ObservedPipelineVersion).To(Equal(src.Status.ObservedPipelineVersion))
		})
	})

	var _ = Describe("ConvertFrom", func() {

		Specify("Converts RuntimeParameters to a map", func() {
			src := v1alpha3.RunConfiguration{Spec: v1alpha3.RunConfigurationSpec{
				RuntimeParameters: []apis.NamedValue{
					{Name: "a", Value: "b"},
					{Name: "c", Value: "d"},
				}}}

			expected := RunConfiguration{Spec: RunConfigurationSpec{
				RuntimeParameters: map[string]string{"a": "b", "c": "d"},
			}}

			dst := RunConfiguration{}
			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst).To(Equal(expected))
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

		Specify("Copies all other fields", func() {
			src := v1alpha3.RunConfiguration{
				ObjectMeta: v1.ObjectMeta{Name: "runconfiguration"},
				Status: v1alpha3.RunConfigurationStatus{
					Status: apis.Status{
						Version:              "1",
						KfpId:                "id",
						SynchronizationState: apis.Succeeded,
					},
					ObservedPipelineVersion: "1",
				},
			}

			dst := RunConfiguration{}
			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.ObjectMeta).To(Equal(src.ObjectMeta))
			Expect(dst.Status.Status).To(Equal(src.Status.Status))
			Expect(dst.Status.ObservedPipelineVersion).To(Equal(src.Status.ObservedPipelineVersion))
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
