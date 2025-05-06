//go:build unit

package v1alpha6

import (
	"encoding/json"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Context("Provider Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip forward", func() {
		Specify("converts to and from the same object", func() {
			srcProvider := RandomProvider()

			srcProvider.Spec.DefaultBeamArgs = apis.RandomNonEmptyList(apis.RandomNamedValue)

			intermediate := &hub.Provider{}
			dst := &Provider{}

			Expect(srcProvider.ConvertTo(intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())
			Expect(dst).To(BeComparableTo(srcProvider, cmpopts.EquateEmpty(), syncStateComparer))
		})

		Specify("converts to and from the same object when there are no default beam args", func() {
			srcProvider := RandomProvider()
			srcProvider.Spec.DefaultBeamArgs = nil

			intermediate := &hub.Provider{}
			dst := &Provider{}

			Expect(srcProvider.ConvertTo(intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())
			Expect(dst).To(BeComparableTo(srcProvider, cmpopts.EquateEmpty(), syncStateComparer))
		})

	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object when tfx is supported", func() {
			src := hub.RandomProvider()
			framework := hub.RandomFramework()
			DefaultTfxImage = framework.Image
			framework.Name = "tfx"
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
		Specify("ConvertFrom fails when there are no annotations and tfx isn't in the supported frameworks", func() {
			src := hub.RandomProvider()
			src.Spec.Frameworks = nil
			dst := Provider{}

			Expect(dst.ConvertFrom(src)).To(Not(Succeed()))
		})
	})
})
