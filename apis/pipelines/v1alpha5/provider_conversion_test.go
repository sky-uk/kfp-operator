//go:build unit

package v1alpha5

import (
	"encoding/json"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
)

var _ = Context("Provider Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomProvider()
			framework := hub.RandomFramework()
			framework.Name = "tfx"
			patchOps := []apis.JsonPatchOperation{
				{
					Op:   "add",
					Path: "/framework/parameters/beamArgs/-",
					Value: map[string]string{
						"name":  "foo",
						"value": "bar",
					},
				},
			}
			bytes, err := json.Marshal(patchOps)
			Expect(err).To(Not(HaveOccurred()))
			framework.Patches = []hub.Patch{
				{Type: "json", Patch: string(bytes)},
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
})
