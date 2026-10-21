//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"sigs.k8s.io/yaml"
)

var _ = Describe("LabelService", func() {
	Context("InsertLabelsIntoParameters", func() {
		When("the pipeline is valid", func() {
			It("inserts each label as a parameter at the configured path", func() {
				labels := []string{"label1", "label2"}
				result, err := DefaultLabelService{
					parameterDefaults: map[string]any{"test": "test"},
					parameterPath:     []string{"testParamPath"},
				}.InsertLabelsIntoParameters(resource.CompiledPipeline(`{}`), labels)
				Expect(err).ToNot(HaveOccurred())

				var got map[string]any
				Expect(yaml.Unmarshal([]byte(result), &got)).To(Succeed())
				Expect(got["testParamPath"]).To(Equal(map[string]any{
					"label1": map[string]any{"test": "test"},
					"label2": map[string]any{"test": "test"},
				}))
			})

			It("creates missing intermediate maps along the path", func() {
				result, err := NewDefaultLabelService().InsertLabelsIntoParameters(resource.CompiledPipeline(`{}`), []string{"label1"})
				Expect(err).NotTo(HaveOccurred())

				var got map[string]any
				Expect(yaml.Unmarshal([]byte(result), &got)).To(Succeed())
				parameters := got["root"].(map[string]any)["inputDefinitions"].(map[string]any)["parameters"].(map[string]any)
				Expect(parameters).To(HaveKey("label1"))
			})

			It("preserves a trailing PlatformSpec document", func() {
				compiled := resource.CompiledPipeline(`root:
  inputDefinitions:
    parameters: {}
---
platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec:
          nodeSelector:
            disktype: ssd
`)
				result, err := NewDefaultLabelService().InsertLabelsIntoParameters(compiled, []string{"label1"})
				Expect(err).NotTo(HaveOccurred())

				documents, err := splitYAMLDocuments([]byte(result))
				Expect(err).NotTo(HaveOccurred())
				Expect(documents).To(HaveLen(2))

				var spec map[string]any
				Expect(yaml.Unmarshal(documents[0], &spec)).To(Succeed())
				parameters := spec["root"].(map[string]any)["inputDefinitions"].(map[string]any)["parameters"].(map[string]any)
				Expect(parameters).To(HaveKey("label1"))

				var platform map[string]any
				Expect(yaml.Unmarshal(documents[1], &platform)).To(Succeed())
				Expect(platform).To(HaveKey("platforms"))
			})
		})

		When("the pipeline is empty", func() {
			It("returns an error", func() {
				_, err := DefaultLabelService{}.InsertLabelsIntoParameters(resource.CompiledPipeline{}, []string{"label1"})
				Expect(err).To(MatchError(ContainSubstring("compiled pipeline is empty")))
			})
		})

		When("a value along the parameter path is not a map", func() {
			It("returns an error", func() {
				result, err := DefaultLabelService{
					parameterDefaults: map[string]any{},
					parameterPath:     []string{"root", "inputDefinitions"},
				}.InsertLabelsIntoParameters(resource.CompiledPipeline("root: not-a-map\n"), []string{"label1"})
				Expect(err).To(MatchError(ContainSubstring(`expected map at key "root"`)))
				Expect(result).To(BeNil())
			})
		})
	})
})
