//go:build unit

package provider

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	yamlv3 "go.yaml.in/yaml/v3"
	"google.golang.org/protobuf/proto"
)

const pipelineSpecYAML = `# PIPELINE DEFINITION
# Name: training
components:
  comp-train:
    executorLabel: exec-train
deploymentSpec:
  executors:
    exec-train:
      container:
        command:
        - python
        image: trainer
pipelineInfo:
  name: training
root:
  dag:
    tasks:
      train:
        componentRef:
          name: comp-train
        taskInfo:
          name: train
  inputDefinitions:
    parameters:
      epochs:
        defaultValue: 5.0
        isOptional: true
        parameterType: NUMBER_INTEGER
schemaVersion: 2.1.0
sdkVersion: kfp-2.17.0
`

const platformSpecYAML = `platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec-train:
          activeDeadlineSeconds: '600'
          securityContext:
            runAsNonRoot: true
            runAsUser: '1000'
          tolerations:
          - key: dedicated
            operator: Exists
`

// decodeYAMLDocuments decodes every document with yaml.v3, as the Kubeflow
// Pipelines API server does.
func decodeYAMLDocuments(compiled resource.CompiledPipeline) []map[string]any {
	decoder := yamlv3.NewDecoder(bytes.NewReader(compiled))
	var documents []map[string]any
	for {
		var document map[string]any
		err := decoder.Decode(&document)
		if errors.Is(err, io.EOF) {
			return documents
		}
		Expect(err).ToNot(HaveOccurred())
		documents = append(documents, document)
	}
}

var _ = Describe("PipelinePackage", func() {
	Context("ParsePipelinePackage", func() {
		It("parses a pipeline spec without a platform spec", func() {
			pipelinePackage, err := ParsePipelinePackage(resource.CompiledPipeline(pipelineSpecYAML), "kfpsdk")

			Expect(err).ToNot(HaveOccurred())
			Expect(pipelinePackage.PipelineSpec.GetPipelineInfo().GetName()).To(Equal("training"))
			Expect(pipelinePackage.PipelineSpec.GetRoot().GetInputDefinitions().GetParameters()).To(HaveKey("epochs"))
			Expect(pipelinePackage.PlatformSpec).To(BeNil())
		})

		It("parses a pipeline spec followed by a platform spec", func() {
			compiled := resource.CompiledPipeline(pipelineSpecYAML + "---\n" + platformSpecYAML)

			pipelinePackage, err := ParsePipelinePackage(compiled, "kfpsdk")

			Expect(err).ToNot(HaveOccurred())
			Expect(pipelinePackage.PipelineSpec.GetPipelineInfo().GetName()).To(Equal("training"))
			Expect(pipelinePackage.KubernetesExecutorConfig("exec-train").AsMap()).To(HaveKey("securityContext"))
		})

		It("unwraps the PipelineJob that TFX emits, for any casing of the framework name", func() {
			var pipelineSpec map[string]any
			Expect(yamlv3.Unmarshal([]byte(pipelineSpecYAML), &pipelineSpec)).To(Succeed())
			compiled, err := json.Marshal(map[string]any{
				"displayName":   "training",
				"pipelineSpec":  pipelineSpec,
				"runtimeConfig": map[string]any{},
			})
			Expect(err).ToNot(HaveOccurred())

			pipelinePackage, err := ParsePipelinePackage(compiled, "TfX")

			Expect(err).ToNot(HaveOccurred())
			Expect(pipelinePackage.PipelineSpec.GetPipelineInfo().GetName()).To(Equal("training"))
		})

		It("accepts a TFX pipeline spec that is not wrapped", func() {
			pipelinePackage, err := ParsePipelinePackage(resource.CompiledPipeline(pipelineSpecYAML), "tfx")

			Expect(err).ToNot(HaveOccurred())
			Expect(pipelinePackage.PipelineSpec.GetPipelineInfo().GetName()).To(Equal("training"))
		})

		It("does not unwrap pipelines of other frameworks", func() {
			_, err := ParsePipelinePackage(resource.CompiledPipeline("pipelineSpec:\n  pipelineInfo:\n    name: training\n"), "kfpsdk")

			Expect(err).To(MatchError(ContainSubstring("invalid pipeline spec")))
		})

		DescribeTable("rejects compiler output the Kubeflow Pipelines API server would reject",
			func(compiled string, expectedErr string) {
				pipelinePackage, err := ParsePipelinePackage(resource.CompiledPipeline(compiled), "kfpsdk")

				Expect(err).To(MatchError(ContainSubstring(expectedErr)))
				Expect(pipelinePackage).To(BeNil())
			},
			Entry("empty", "", "compiled pipeline is empty"),
			Entry("only comments", "# PIPELINE DEFINITION\n", "compiled pipeline is empty"),
			Entry("invalid YAML", "root: [unclosed", "failed to decode compiled pipeline"),
			Entry("not a mapping", "- a\n- b\n", "invalid pipeline spec"),
			Entry("unknown pipeline spec field", "unknownField: true\n", "invalid pipeline spec"),
			Entry("unknown platform spec field", pipelineSpecYAML+"---\nunknownField: true\n", "invalid platform spec"),
			Entry("more than two documents", pipelineSpecYAML+"---\n"+platformSpecYAML+"---\n"+platformSpecYAML, "got 3 documents"),
		)
	})

	Context("Marshal", func() {
		It("writes a document that parses back to the same pipeline and platform spec", func() {
			original, err := ParsePipelinePackage(resource.CompiledPipeline(pipelineSpecYAML+"---\n"+platformSpecYAML), "kfpsdk")
			Expect(err).ToNot(HaveOccurred())

			marshalled, err := original.Marshal()
			Expect(err).ToNot(HaveOccurred())
			Expect(decodeYAMLDocuments(marshalled)).To(HaveLen(2))

			roundTripped, err := ParsePipelinePackage(marshalled, "kfpsdk")
			Expect(err).ToNot(HaveOccurred())
			Expect(proto.Equal(roundTripped.PipelineSpec, original.PipelineSpec)).To(BeTrue())
			Expect(proto.Equal(roundTripped.PlatformSpec, original.PlatformSpec)).To(BeTrue())
		})

		It("keeps Kubernetes executor settings it has no part in", func() {
			pipelinePackage, err := ParsePipelinePackage(resource.CompiledPipeline(pipelineSpecYAML+"---\n"+platformSpecYAML), "kfpsdk")
			Expect(err).ToNot(HaveOccurred())

			marshalled, err := pipelinePackage.Marshal()
			Expect(err).ToNot(HaveOccurred())

			documents := decodeYAMLDocuments(marshalled)
			var expectedPlatformSpec map[string]any
			Expect(yamlv3.Unmarshal([]byte(platformSpecYAML), &expectedPlatformSpec)).To(Succeed())
			Expect(documents[1]).To(Equal(expectedPlatformSpec))
		})

		It("writes a single document when there is no platform spec", func() {
			pipelinePackage, err := ParsePipelinePackage(resource.CompiledPipeline(pipelineSpecYAML), "kfpsdk")
			Expect(err).ToNot(HaveOccurred())

			marshalled, err := pipelinePackage.Marshal()
			Expect(err).ToNot(HaveOccurred())
			Expect(decodeYAMLDocuments(marshalled)).To(HaveLen(1))
		})

		It("writes scripts that start with whitespace so they parse back unchanged", func() {
			scripts := []any{
				"\nif ! [ -x \"$(command -v pip)\" ]; then\n    python3 -m ensurepip\nfi\n",
				"\n\nimport kfp\n\n    def f():\n        pass\n\n\n",
				"  indented first line\nsecond line\n",
			}
			command, err := json.Marshal(append([]any{"sh", "-c"}, scripts...))
			Expect(err).ToNot(HaveOccurred())
			compiled := resource.CompiledPipeline(`{"pipelineInfo": {"name": "training"}, "root": {}, "deploymentSpec": {"executors": {"exec-train": {"container": {"image": "trainer", "command": ` + string(command) + `}}}}}`)

			pipelinePackage, err := ParsePipelinePackage(compiled, "kfpsdk")
			Expect(err).ToNot(HaveOccurred())
			marshalled, err := pipelinePackage.Marshal()
			Expect(err).ToNot(HaveOccurred())

			documents := decodeYAMLDocuments(marshalled)
			Expect(documents).To(HaveLen(1))
			container := documents[0]["deploymentSpec"].(map[string]any)["executors"].(map[string]any)["exec-train"].(map[string]any)["container"].(map[string]any)
			Expect(container["command"]).To(Equal(append([]any{"sh", "-c"}, scripts...)))
		})
	})
})
