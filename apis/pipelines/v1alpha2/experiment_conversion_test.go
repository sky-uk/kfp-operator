package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Context("Experiment Conversion", func() {
	var _ = Describe("ConvertTo", func() {

		Specify("Copies all fields", func() {
			src := Experiment{
				ObjectMeta: v1.ObjectMeta{Name: "experiment"},
				Spec: ExperimentSpec{
					Description: "experiment description",
				},
				Status: apis.Status{
					Version:              "1",
					KfpId:                "id",
					SynchronizationState: apis.Succeeded,
				},
			}

			dst := v1alpha3.Experiment{}
			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst.ObjectMeta).To(Equal(src.ObjectMeta))
			Expect(dst.Spec.Description).To(Equal(src.Spec.Description))
			Expect(dst.Status).To(Equal(src.Status))
		})
	})

	var _ = Describe("ConvertFrom", func() {

		Specify("Copies all fields", func() {
			src := v1alpha3.Experiment{
				ObjectMeta: v1.ObjectMeta{Name: "experiment"},
				Spec: v1alpha3.ExperimentSpec{
					Description: "experiment description",
				},
				Status: apis.Status{
					Version:              "1",
					KfpId:                "id",
					SynchronizationState: apis.Succeeded,
				},
			}

			dst := Experiment{}
			Expect(dst.ConvertFrom(&src)).To(Succeed())
			Expect(dst.ObjectMeta).To(Equal(src.ObjectMeta))
			Expect(dst.Spec.Description).To(Equal(src.Spec.Description))
			Expect(dst.Status).To(Equal(src.Status))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Does not change between versions", func() {
			src := Experiment{
				Spec: ExperimentSpec{Description: "description"},
			}

			dst := v1alpha3.Experiment{}
			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(src.Spec.ComputeVersion()).To(Equal(dst.Spec.ComputeVersion()))
		})
	})
})
