//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	sigsyaml "sigs.k8s.io/yaml"
)

var _ = Describe("LabelService", func() {
	Context("InsertLabelsIntoParameters", func() {
		When("given a list of labels keys and a compiled pipeline", func() {
			It("inserts the labels with a default value into the compiled pipeline", func() {
				stubPipelineBytes := []byte(`{}`)
				stubParamDefaultValue := []byte(`{"test": "test"}`)

				labels := []string{"label1", "label2"}

				result, err := DefaultLabelService{
					parameterDefaults: stubParamDefaultValue,
					parameterJsonPath: "testParamPath",
				}.InsertLabelsIntoParameters(stubPipelineBytes, labels)
				Expect(err).ToNot(HaveOccurred())

				var got map[string]any
				Expect(sigsyaml.Unmarshal(result, &got)).To(Succeed())
				Expect(got["testParamPath"]).To(Equal(map[string]any{
					"label1": map[string]any{"test": "test"},
					"label2": map[string]any{"test": "test"},
				}))
			})

			It("keeps a following platformSpec document", func() {
				compiled := []byte("root:\n  inputDefinitions:\n    parameters: {}\n---\nplatforms:\n  kubernetes:\n    deploymentSpec:\n      executors:\n        exec:\n          nodeSelector:\n            disktype: ssd\n")
				service, err := NewDefaultLabelService()
				Expect(err).NotTo(HaveOccurred())

				result, err := service.InsertLabelsIntoParameters(compiled, []string{"label1"})
				Expect(err).NotTo(HaveOccurred())

				documents, err := splitYAMLDocuments(result)
				Expect(err).NotTo(HaveOccurred())
				Expect(documents).To(HaveLen(2))

				var spec map[string]any
				Expect(sigsyaml.Unmarshal(documents[0], &spec)).To(Succeed())
				parameters := spec["root"].(map[string]any)["inputDefinitions"].(map[string]any)["parameters"].(map[string]any)
				Expect(parameters).To(HaveKey("label1"))

				var platform map[string]any
				Expect(sigsyaml.Unmarshal(documents[1], &platform)).To(Succeed())
				Expect(platform).To(HaveKey("platforms"))
			})
		})
	})
})
