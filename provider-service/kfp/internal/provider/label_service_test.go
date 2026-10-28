//go:build unit

package provider

import (
	"io"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	yamlv3 "go.yaml.in/yaml/v3"
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

			It("replaces an existing entry with the same name", func() {
				compiled := resource.CompiledPipeline(`root:
  inputDefinitions:
    parameters:
      label1:
        defaultValue: old
        parameterType: NUMBER_INTEGER
`)
				result, err := NewDefaultLabelService().InsertLabelsIntoParameters(compiled, []string{"label1"})
				Expect(err).NotTo(HaveOccurred())

				var got map[string]any
				Expect(yaml.Unmarshal([]byte(result), &got)).To(Succeed())
				parameters := got["root"].(map[string]any)["inputDefinitions"].(map[string]any)["parameters"].(map[string]any)
				Expect(parameters["label1"]).To(Equal(map[string]any{
					"isOptional":    true,
					"parameterType": "STRING",
					"defaultValue":  "",
				}))
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

				decoder := yamlv3.NewDecoder(strings.NewReader(string(result)))
				var spec, platform, extra map[string]any
				Expect(decoder.Decode(&spec)).To(Succeed())
				Expect(decoder.Decode(&platform)).To(Succeed())
				Expect(decoder.Decode(&extra)).To(MatchError(io.EOF))

				parameters := spec["root"].(map[string]any)["inputDefinitions"].(map[string]any)["parameters"].(map[string]any)
				Expect(parameters).To(HaveKey("label1"))
				Expect(platform).To(HaveKey("platforms"))
			})

			It("preserves comments, number formatting and quoted strings", func() {
				compiled := resource.CompiledPipeline(`# PIPELINE DEFINITION
root:
  inputDefinitions:
    parameters:
      a:
        defaultValue: 1.0
        parameterType: NUMBER_DOUBLE
deploymentSpec:
  executors:
    exec:
      container:
        command:
        - sh
        - "\nif ! [ -x ]; then\n    python3\nfi\n"
`)
				result, err := NewDefaultLabelService().InsertLabelsIntoParameters(compiled, []string{"label1"})
				Expect(err).NotTo(HaveOccurred())

				Expect(string(result)).To(ContainSubstring("# PIPELINE DEFINITION"))
				Expect(string(result)).To(ContainSubstring("defaultValue: 1.0"))
				Expect(string(result)).To(ContainSubstring(`"\nif ! [ -x ]; then\n    python3\nfi\n"`))
			})

			It("writes block scalars that start with whitespace so they parse back unchanged", func() {
				scripts := []string{
					"\nif ! [ -x \"$(command -v pip)\" ]; then\n    python3 -m ensurepip\nfi\n",
					"\n\nimport kfp\n\n    def f():\n        pass\n\n\n",
					"  indented first line\nsecond line\n",
				}
				command := []any{"sh", "-c"}
				for _, script := range scripts {
					command = append(command, script)
				}
				compiled, err := yaml.Marshal(map[string]any{
					"deploymentSpec": map[string]any{"executors": map[string]any{
						"exec": map[string]any{"container": map[string]any{"command": command}},
					}},
				})
				Expect(err).NotTo(HaveOccurred())

				result, err := NewDefaultLabelService().InsertLabelsIntoParameters(resource.CompiledPipeline(compiled), []string{"label1"})
				Expect(err).NotTo(HaveOccurred())

				// Parse with yaml.v3, as the KFP API server does.
				var got map[string]any
				Expect(yamlv3.Unmarshal([]byte(result), &got)).To(Succeed())
				executor := got["deploymentSpec"].(map[string]any)["executors"].(map[string]any)["exec"].(map[string]any)
				Expect(executor["container"].(map[string]any)["command"]).To(Equal(command))
			})
		})

		When("the pipeline is empty", func() {
			It("returns an error", func() {
				_, err := DefaultLabelService{}.InsertLabelsIntoParameters(resource.CompiledPipeline{}, []string{"label1"})
				Expect(err).To(MatchError(ContainSubstring("compiled pipeline is empty")))
			})
		})

		When("the pipeline is not valid YAML", func() {
			It("returns an error", func() {
				_, err := NewDefaultLabelService().InsertLabelsIntoParameters(resource.CompiledPipeline("root: [unclosed"), []string{"label1"})
				Expect(err).To(MatchError(ContainSubstring("failed to decode compiled pipeline")))
			})
		})

		When("the pipeline is not a mapping", func() {
			It("returns an error", func() {
				_, err := NewDefaultLabelService().InsertLabelsIntoParameters(resource.CompiledPipeline("- a\n- b\n"), []string{"label1"})
				Expect(err).To(MatchError(ContainSubstring("expected a mapping at the document root")))
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
