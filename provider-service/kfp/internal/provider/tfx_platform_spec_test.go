//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"sigs.k8s.io/yaml"
)

var _ = Describe("addHackedTFXPlatformSpec", func() {
	pipeline := resource.CompiledPipeline(`pipelineInfo:
  name: p
deploymentSpec:
  executors:
    exec-importer:
      importer:
        artifactUri:
          constant: x
    exec-trainer:
      container:
        image: img
    exec-example-gen:
      container:
        image: img
root:
  dag: {}
---
platforms:
  kubernetes:
    deploymentSpec:
      executors:
        stale: {}
`)

	It("adds a configMapAsVolume to every container executor only, replacing any existing platform spec", func() {
		result, err := addHackedTFXPlatformSpec(pipeline, "TFX")
		Expect(err).ToNot(HaveOccurred())

		documents, err := decodeYAMLDocuments([]byte(result))
		Expect(err).ToNot(HaveOccurred())
		Expect(documents).To(HaveLen(2))

		var platformSpec map[string]any
		Expect(yaml.Unmarshal([]byte(result[indexOfSecondDocument(result):]), &platformSpec)).To(Succeed())

		executors := platformSpec["platforms"].(map[string]any)["kubernetes"].(map[string]any)["deploymentSpec"].(map[string]any)["executors"].(map[string]any)
		Expect(executors).To(HaveLen(2))
		Expect(executors).To(HaveKey("exec-trainer"))
		Expect(executors).To(HaveKey("exec-example-gen"))
		Expect(executors["exec-trainer"]).To(Equal(map[string]any{
			"configMapAsVolume": []any{
				map[string]any{
					"configMapName": hackConfigMapName,
					"mountPath":     hackMountPath,
					"optional":      false,
				},
			},
		}))
	})

	It("preserves the pipeline spec as the first document", func() {
		result, err := addHackedTFXPlatformSpec(pipeline, "tfx")
		Expect(err).ToNot(HaveOccurred())

		documents, err := decodeYAMLDocuments([]byte(result))
		Expect(err).ToNot(HaveOccurred())

		var first map[string]any
		Expect(documents[0].Decode(&first)).To(Succeed())
		Expect(first).To(HaveKey("deploymentSpec"))
		Expect(first).To(HaveKey("root"))
	})

	It("uses a ConfigMap name that is a valid pod volume name (lowercase DNS-1123 label)", func() {
		Expect(hackConfigMapName).To(MatchRegexp(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`))
	})

	It("leaves non-TFX pipelines untouched", func() {
		result, err := addHackedTFXPlatformSpec(pipeline, "kfpsdk")
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(pipeline))
	})

	It("is idempotent", func() {
		once, err := addHackedTFXPlatformSpec(pipeline, "tfx")
		Expect(err).ToNot(HaveOccurred())
		twice, err := addHackedTFXPlatformSpec(once, "tfx")
		Expect(err).ToNot(HaveOccurred())
		Expect(twice).To(Equal(once))
	})
})

func indexOfSecondDocument(pipeline resource.CompiledPipeline) int {
	s := string(pipeline)
	for i := 0; i+4 <= len(s); i++ {
		if s[i:i+4] == "---\n" && (i == 0 || s[i-1] == '\n') {
			return i + 4
		}
	}
	return 0
}
