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

			AddComponentsToFrameworkParams("", &framework)

			Expect(framework.Parameters).To(HaveKey("components"))
		})

		Specify("adds tfx framework for some string", func() {
			tfxComponents := apis.RandomString()
			framework := NewPipelineFramework("tfx")

			AddComponentsToFrameworkParams(tfxComponents, &framework)

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
			Expect(components).To(Equal(""))
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
			Expect(components).To(Equal(""))
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
})
