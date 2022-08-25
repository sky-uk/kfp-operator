package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
)

var _ = Context("Pipeline Conversion", func() {
	var _ = Describe("ConvertTo", func() {

		Specify("Converts Env to a list of NamedValue", func() {
			src := Pipeline{Spec: PipelineSpec{Env: map[string]string{"a": "b", "c": "d"}}}
			expected := v1alpha3.Pipeline{Spec: v1alpha3.PipelineSpec{Env: []v1alpha3.NamedValue{
				{Name: "a", Value: "b"},
				{Name: "c", Value: "d"},
			}}}

			dst := v1alpha3.Pipeline{}
			Expect(src.ConvertTo(&dst)).To(Succeed())
			Expect(dst).To(Equal(expected))
		})

	})
})
