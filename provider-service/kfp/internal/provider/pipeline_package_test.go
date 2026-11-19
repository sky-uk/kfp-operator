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

func decodeYAMLDocuments(compiled resource.CompiledPipeline) []map[string]any {
	GinkgoHelper()

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
		DescribeTable("parses a pipeline spec, unwrapping the PipelineJob TFX emits",
			func(compiled string, frameworkName string) {
				pipelinePackage, err := ParsePipelinePackage(resource.CompiledPipeline(compiled), frameworkName)

				Expect(err).ToNot(HaveOccurred())
				Expect(pipelinePackage.PipelineSpec.GetPipelineInfo().GetName()).To(Equal("training"))
				Expect(pipelinePackage.PlatformSpec).To(BeNil())
			},
			Entry("KFP SDK", pipelineSpecYAML, "kfpsdk"),
			Entry("TFX without a PipelineJob", pipelineSpecYAML, "tfx"),
			Entry("TFX with a PipelineJob, in any casing", `{"pipelineSpec": {"pipelineInfo": {"name": "training"}, "root": {}}}`, "TfX"),
		)

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
			Entry("a PipelineJob from another framework", "pipelineSpec:\n  pipelineInfo:\n    name: training\n", "invalid pipeline spec"),
			Entry("unknown pipeline spec field", "unknownField: true\n", "invalid pipeline spec"),
			Entry("unknown platform spec field", pipelineSpecYAML+"---\nunknownField: true\n", "invalid platform spec"),
			Entry("more than two documents", pipelineSpecYAML+"---\n"+platformSpecYAML+"---\n"+platformSpecYAML, "got 3 documents"),
		)
	})

	Context("Marshal", func() {
		It("writes the pipeline and platform spec so they parse back unchanged", func() {
			original, err := ParsePipelinePackage(resource.CompiledPipeline(pipelineSpecYAML+"---\n"+platformSpecYAML), "kfpsdk")
			Expect(err).ToNot(HaveOccurred())

			marshalled, err := original.Marshal()
			Expect(err).ToNot(HaveOccurred())

			var platformSpec map[string]any
			Expect(yamlv3.Unmarshal([]byte(platformSpecYAML), &platformSpec)).To(Succeed())
			documents := decodeYAMLDocuments(marshalled)
			Expect(documents).To(HaveLen(2))
			Expect(documents[1]).To(Equal(platformSpec))

			roundTripped, err := ParsePipelinePackage(marshalled, "kfpsdk")
			Expect(err).ToNot(HaveOccurred())
			Expect(proto.Equal(roundTripped.PipelineSpec, original.PipelineSpec)).To(BeTrue())
		})

		It("writes scripts that start with whitespace so they parse back unchanged", func() {
			command := []any{
				"sh",
				"-c",
				"\nif ! [ -x \"$(command -v pip)\" ]; then\n    python3 -m ensurepip\nfi\n",
				"\n\nimport kfp\n\n    def f():\n        pass\n\n\n",
				"  indented first line\nsecond line\n",
			}
			compiled, err := json.Marshal(map[string]any{
				"deploymentSpec": map[string]any{"executors": map[string]any{
					"exec-train": map[string]any{"container": map[string]any{"image": "trainer", "command": command}},
				}},
				"pipelineInfo": map[string]any{"name": "training"},
				"root":         map[string]any{},
			})
			Expect(err).ToNot(HaveOccurred())

			pipelinePackage, err := ParsePipelinePackage(compiled, "kfpsdk")
			Expect(err).ToNot(HaveOccurred())
			marshalled, err := pipelinePackage.Marshal()
			Expect(err).ToNot(HaveOccurred())

			documents := decodeYAMLDocuments(marshalled)
			Expect(documents).To(HaveLen(1))
			Expect(documents[0]).To(HaveKeyWithValue("deploymentSpec", map[string]any{"executors": map[string]any{
				"exec-train": map[string]any{"container": map[string]any{"image": "trainer", "command": command}},
			}}))
		})
	})
})
