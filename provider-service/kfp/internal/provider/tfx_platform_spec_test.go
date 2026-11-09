//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
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

	It("replaces the platform spec with ConfigMap volume and env config for container executors", func() {
		result, err := addHackedTFXPlatformSpec(pipeline, "TFX")
		Expect(err).ToNot(HaveOccurred())

		documents, err := decodeYAMLDocuments([]byte(result))
		Expect(err).ToNot(HaveOccurred())
		Expect(documents).To(HaveLen(2))

		var platformSpec map[string]any
		Expect(documents[1].Decode(&platformSpec)).To(Succeed())

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
			"configMapAsEnv": []any{
				map[string]any{
					"configMapName": hackEnvConfigMapName,
					"keyToEnv": []any{
						map[string]any{
							"configMapKey": hackEnvConfigMapKey,
							"envVar":       hackEnvVarName,
						},
					},
				},
			},
		}))
	})

	It("keeps the pipeline spec as the first document", func() {
		result, err := addHackedTFXPlatformSpec(pipeline, "tfx")
		Expect(err).ToNot(HaveOccurred())

		documents, err := decodeYAMLDocuments([]byte(result))
		Expect(err).ToNot(HaveOccurred())

		var first map[string]any
		Expect(documents[0].Decode(&first)).To(Succeed())
		Expect(first).To(HaveKey("deploymentSpec"))
		Expect(first).To(HaveKey("root"))
	})

	It("uses a DNS-1123 label as the volume ConfigMap name", func() {
		Expect(hackConfigMapName).To(MatchRegexp(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`))
	})

	It("leaves non-TFX pipelines unchanged", func() {
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
