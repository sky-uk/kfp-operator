//go:build unit

package v1alpha5

import (
	"encoding/json"
	"fmt"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Context("Provider Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomProvider()
			framework := hub.RandomFramework()
			framework.Name = "tfx"
			DefaultTfxImage = framework.Image

			patchOps := []apis.JsonPatchOperation{
				{
					Op:   "add",
					Path: "/framework/parameters/beamArgs/0",
					Value: map[string]string{
						"name":  "foo",
						"value": "bar",
					},
				},
			}
			bytes, err := json.Marshal(patchOps)
			Expect(err).To(Not(HaveOccurred()))
			framework.Patches = []hub.Patch{
				{Type: "json", Payload: string(bytes)},
			}

			src.Spec.Frameworks = []hub.Framework{framework}

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

		Specify("ConvertFrom fails when there are no annotations and tfx isn't in the supported frameworks", func() {
			src := hub.RandomProvider()
			src.Spec.Frameworks = nil
			dst := Provider{}

			Expect(dst.ConvertFrom(src)).To(Not(Succeed()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object when the framework is tfx", func() {
			src := RandomProvider()
			remainder := ProviderConversionRemainder{
				ServiceImage: "foo",
			}
			Expect(pipelines.SetConversionAnnotations(src, &remainder)).To(Succeed())

			intermediate := &hub.Provider{}
			dst := &Provider{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			intRemainderStr := intermediate.Annotations[remainder.ConversionAnnotation()]
			intRemainder := ProviderConversionRemainder{}
			Expect(json.Unmarshal([]byte(intRemainderStr), &intRemainder)).To(Succeed())
			Expect(intRemainder.Image).To(Equal(src.Spec.Image))
			Expect(intRemainder.ServiceImage).To(Equal(""))

			Expect(dst.ConvertFrom(intermediate)).To(Succeed())
			Expect(pipelines.SetConversionAnnotations(src, &remainder)).To(Succeed())

			println(fmt.Sprintf("src: %+v", src))
			println(fmt.Sprintf("intermediate: %+v", intermediate))
			println(fmt.Sprintf("dst: %+v", dst))

			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty(), syncStateComparer))
		})
	})
})
