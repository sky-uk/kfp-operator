//go:build unit

package v1beta1

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var _ = Context("Conversions", func() {
	var _ = Describe("AddComponentsToFrameworkParams", func() {
		Specify("adds tfx framework for empty string", func() {
			framework := NewPipelineFramework("tfx")

			err := AddComponentsToFrameworkParams("", &framework)
			Expect(err).To(Not(HaveOccurred()))

			Expect(framework.Parameters).To(HaveKey("components"))
		})

		Specify("adds tfx framework for some string", func() {
			tfxComponents := apis.RandomString()
			framework := NewPipelineFramework("tfx")

			err := AddComponentsToFrameworkParams(tfxComponents, &framework)
			Expect(err).To(Not(HaveOccurred()))

			marshal, _ := json.Marshal(tfxComponents)
			Expect(framework.Parameters["components"]).To(Equal(&apiextensionsv1.JSON{Raw: marshal}))
		})
	})

	var _ = Describe("AddBeamArgsToFrameworkParams", func() {
		Specify("adds beamArgs to framework params for empty list", func() {
			framework := NewPipelineFramework("tfx")

			err := AddBeamArgsToFrameworkParams([]apis.NamedValue{}, &framework)

			Expect(err).To(Not(HaveOccurred()))
			Expect(framework.Parameters).To(HaveKey("beamArgs"))
		})

		Specify("adds beamArgs to framework params for populated list", func() {
			beamArgs := []apis.NamedValue{
				{Name: "name1", Value: "value1"},
				{Name: "name2", Value: "value2"},
			}

			framework := NewPipelineFramework("tfx")

			err := AddBeamArgsToFrameworkParams(beamArgs, &framework)
			Expect(err).To(Not(HaveOccurred()))

			marshal, _ := json.Marshal(beamArgs)
			Expect(framework.Parameters["beamArgs"]).To(Equal(&apiextensionsv1.JSON{Raw: marshal}))
		})
	})

	var _ = Describe("ComponentsFromFramework", func() {
		Specify("returns empty string for empty framework", func() {
			framework := NewPipelineFramework("tfx")
			framework.Parameters = nil

			components, err := ComponentsFromFramework(&framework)

			Expect(err).To(Not(HaveOccurred()))
			Expect(components).To(BeEmpty())
		})

		Specify("returns components for populated framework", func() {
			components := apis.RandomString()

			jsonString, err := json.Marshal(components)
			Expect(err).To(Not(HaveOccurred()))

			framework := NewPipelineFramework("tfx")
			framework.Parameters["components"] = &apiextensionsv1.JSON{Raw: jsonString}

			result, err := ComponentsFromFramework(&framework)

			Expect(err).To(Not(HaveOccurred()))
			Expect(components).To(Equal(result))
		})

		Specify("returns empty string if components do not exist", func() {
			framework := NewPipelineFramework("tfx")

			components, err := ComponentsFromFramework(&framework)

			Expect(err).To(Not(HaveOccurred()))
			Expect(components).To(BeEmpty())
		})
	})

	var _ = Describe("BeamArgsFromFramework", func() {
		Specify("returns empty list for empty framework parameters", func() {
			framework := NewPipelineFramework("tfx")
			framework.Parameters = nil

			beamArgs, err := BeamArgsFromFramework(&framework)

			Expect(err).To(Not(HaveOccurred()))
			Expect(beamArgs).To(BeEmpty())
		})

		Specify("returns beamArgs for populated framework", func() {
			beamArgs := []apis.NamedValue{
				{Name: "name1", Value: "value1"},
				{Name: "name2", Value: "value2"},
			}

			framework := NewPipelineFramework("tfx")

			marshal, _ := json.Marshal(beamArgs)
			framework.Parameters["beamArgs"] = &apiextensionsv1.JSON{Raw: marshal}

			beamArgsFromFramework, err := BeamArgsFromFramework(&framework)

			Expect(err).To(Not(HaveOccurred()))
			Expect(beamArgsFromFramework).To(Equal(beamArgs))
		})

		Specify("returns empty list if beamArgs do not exist", func() {
			framework := NewPipelineFramework("tfx")

			beamArgs, err := BeamArgsFromFramework(&framework)

			Expect(err).To(Not(HaveOccurred()))
			Expect(beamArgs).To(BeEmpty())
		})
	})

	var _ = Describe("BeamArgsFromJsonPatches", func() {
		Specify("returns empty list for empty patches", func() {
			var patches []Patch

			beamArgs, err := BeamArgsFromJsonPatches(patches)

			Expect(err).To(Not(HaveOccurred()))
			Expect(beamArgs).To(BeEmpty())
		})

		Specify("returns beamArgs for populated patches", func() {
			beamArgs := []apis.NamedValue{
				{Name: "name1", Value: "value1"},
				{Name: "name2", Value: "value2"},
			}

			addOp := apis.JsonPatchOperation{
				Op:    "add",
				Path:  "foo",
				Value: map[string]string{"name": "name1", "value": "value1"},
			}

			addOp2 := addOp
			addOp2.Value = map[string]string{"name": "name2", "value": "value2"}

			patchOps := []apis.JsonPatchOperation{addOp, addOp2}
			bytes, err := json.Marshal(patchOps)
			Expect(err).To(Not(HaveOccurred()))
			patches := []Patch{
				{Type: "json", Patch: string(bytes)},
			}

			beamArgsFromPatches, err := BeamArgsFromJsonPatches(patches)

			Expect(err).To(Not(HaveOccurred()))
			Expect(beamArgsFromPatches).To(Equal(beamArgs))
		})

		Specify("fail if a patch contains a value that isn't a map[string]interface{}", func() {
			addOp := apis.JsonPatchOperation{
				Op:    "add",
				Path:  "foo",
				Value: "fail",
			}

			addOp2 := addOp
			addOp2.Value = map[string]string{"name": "name2", "value": "value2"}

			patchOps := []apis.JsonPatchOperation{addOp, addOp2}
			bytes, err := json.Marshal(patchOps)
			Expect(err).To(Not(HaveOccurred()))
			patches := []Patch{
				{Type: "json", Patch: string(bytes)},
			}

			_, err = BeamArgsFromJsonPatches(patches)

			Expect(err).To(HaveOccurred())
		})

		Specify("fail if a patch contains a value that isn't a map[string]string{}", func() {
			addOp := apis.JsonPatchOperation{
				Op:    "add",
				Path:  "foo",
				Value: map[string]int{"name": 1, "value": 2},
			}

			addOp2 := addOp
			addOp2.Value = map[string]string{"name": "name2", "value": "value2"}

			patchOps := []apis.JsonPatchOperation{addOp, addOp2}
			bytes, err := json.Marshal(patchOps)
			Expect(err).To(Not(HaveOccurred()))
			patches := []Patch{
				{Type: "json", Patch: string(bytes)},
			}

			_, err = BeamArgsFromJsonPatches(patches)

			Expect(err).To(HaveOccurred())
		})
	})
})
