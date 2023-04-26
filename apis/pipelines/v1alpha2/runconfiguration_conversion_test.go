//go:build unit
// +build unit

package v1alpha2

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("RunConfiguration Conversion", func() {
	var _ = Describe("ConvertTo", func() {
		Specify("Converts RuntimeParameters to a list of NamedValue", func() {
			src := RunConfiguration{Spec: RunConfigurationSpec{RuntimeParameters: map[string]string{"a": "b", "c": "d"}}}
			dst := hub.RunConfiguration{}

			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst.Spec.Run.RuntimeParameters).To(Equal([]apis.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}))
		})
	})

	var _ = Describe("ConvertFrom", func() {
		Specify("Converts RuntimeParameters to a map", func() {
			src := hub.RunConfiguration{Spec: hub.RunConfigurationSpec{Run: hub.RunSpec{
				RuntimeParameters: []apis.NamedValue{
					{Name: "a", Value: "b"},
					{Name: "c", Value: "d"},
				}}}}
			dst := RunConfiguration{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.Spec.RuntimeParameters).To(Equal(map[string]string{"a": "b", "c": "d"}))
		})

		Specify("Removes duplicates and adds remainder annotation when RuntimeParameters contains a duplicate NamedValue", func() {
			src := hub.RunConfiguration{Spec: hub.RunConfigurationSpec{Run: hub.RunSpec{
				RuntimeParameters: []apis.NamedValue{
					{Name: "a", Value: "b"},
					{Name: "a", Value: "d"},
				}}}}
			dst := RunConfiguration{}

			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.Spec.RuntimeParameters).To(Equal(map[string]string{"a": "b"}))
		})
	})

	var _ = Describe("Roundtrip forward", func() {
		Specify("converts to and from the same object", func() {
			src := RandomRunConfiguration()
			src.Status.KfpId = ""
			src.Status.Version = ""
			intermediate := &hub.RunConfiguration{}
			dst := &RunConfiguration{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunConfiguration()
			src.Spec.Triggers = []hub.Trigger{hub.RandomCronTrigger()}

			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())

			Expect(src.Spec.Run.RuntimeParameters).To(ConsistOf(dst.Spec.Run.RuntimeParameters))
			src.Spec.Run.RuntimeParameters = nil
			dst.Spec.Run.RuntimeParameters = nil

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})

		Specify("Duplicate entries are preserved on the roundtrip", func() {
			src := hub.RunConfiguration{
				ObjectMeta: v1.ObjectMeta{
					Annotations: map[string]string{},
				},
				Spec: hub.RunConfigurationSpec{
					Run: hub.RunSpec{
						RuntimeParameters: []apis.NamedValue{
							{Name: "a", Value: "b"},
							{Name: "a", Value: "d"},
						},
					},
				},
			}

			intermediate := RunConfiguration{}
			dst := hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(&src)).To(Succeed())
			Expect(intermediate.ConvertTo(&dst)).To(Succeed())

			Expect(src).To(Equal(dst))
		})
	})
})
