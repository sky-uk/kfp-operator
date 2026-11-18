//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

var _ = Describe("addHackedTFXPlatformSpec", func() {
	const (
		volumeConfigMap = "test-volume-config"
		volumeMountPath = "/etc/test-mount"
		envConfigMap    = "test-env-config"
		envKey          = "TEST_ENV_VAR"
	)

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

	setAll := func() {
		GinkgoT().Setenv(volumeConfigMapEnvVar, volumeConfigMap)
		GinkgoT().Setenv(volumeMountPathEnvVar, volumeMountPath)
		GinkgoT().Setenv(envConfigMapEnvVar, envConfigMap)
		GinkgoT().Setenv(envKeyEnvVar, envKey)
	}

	executorsOf := func(result resource.CompiledPipeline) map[string]any {
		documents, err := decodeYAMLDocuments([]byte(result))
		Expect(err).ToNot(HaveOccurred())
		Expect(documents).To(HaveLen(2))

		var platformSpec map[string]any
		Expect(documents[1].Decode(&platformSpec)).To(Succeed())

		return platformSpec["platforms"].(map[string]any)["kubernetes"].(map[string]any)["deploymentSpec"].(map[string]any)["executors"].(map[string]any)
	}

	It("replaces the platform spec with ConfigMap volume and env config for container executors", func() {
		setAll()

		result, err := addHackedTFXPlatformSpec(pipeline, "TFX")
		Expect(err).ToNot(HaveOccurred())

		executors := executorsOf(result)
		Expect(executors).To(HaveLen(2))
		Expect(executors).To(HaveKey("exec-trainer"))
		Expect(executors).To(HaveKey("exec-example-gen"))
		Expect(executors["exec-trainer"]).To(Equal(map[string]any{
			"configMapAsVolume": []any{
				map[string]any{
					"configMapName": volumeConfigMap,
					"mountPath":     volumeMountPath,
					"optional":      false,
				},
			},
			"configMapAsEnv": []any{
				map[string]any{
					"configMapName": envConfigMap,
					"keyToEnv": []any{
						map[string]any{
							"configMapKey": envKey,
							"envVar":       envKey,
						},
					},
				},
			},
		}))
	})

	It("only adds the volume when only the volume settings are configured", func() {
		GinkgoT().Setenv(volumeConfigMapEnvVar, volumeConfigMap)
		GinkgoT().Setenv(volumeMountPathEnvVar, volumeMountPath)

		result, err := addHackedTFXPlatformSpec(pipeline, "tfx")
		Expect(err).ToNot(HaveOccurred())

		executor := executorsOf(result)["exec-trainer"].(map[string]any)
		Expect(executor).To(HaveKey("configMapAsVolume"))
		Expect(executor).ToNot(HaveKey("configMapAsEnv"))
	})

	It("only adds the env var when only the env settings are configured", func() {
		GinkgoT().Setenv(envConfigMapEnvVar, envConfigMap)
		GinkgoT().Setenv(envKeyEnvVar, envKey)

		result, err := addHackedTFXPlatformSpec(pipeline, "tfx")
		Expect(err).ToNot(HaveOccurred())

		executor := executorsOf(result)["exec-trainer"].(map[string]any)
		Expect(executor).To(HaveKey("configMapAsEnv"))
		Expect(executor).ToNot(HaveKey("configMapAsVolume"))
	})

	It("keeps the pipeline spec as the first document", func() {
		setAll()

		result, err := addHackedTFXPlatformSpec(pipeline, "tfx")
		Expect(err).ToNot(HaveOccurred())

		documents, err := decodeYAMLDocuments([]byte(result))
		Expect(err).ToNot(HaveOccurred())

		var first map[string]any
		Expect(documents[0].Decode(&first)).To(Succeed())
		Expect(first).To(HaveKey("deploymentSpec"))
		Expect(first).To(HaveKey("root"))
	})

	It("leaves the pipeline unchanged when nothing is configured", func() {
		result, err := addHackedTFXPlatformSpec(pipeline, "tfx")
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(pipeline))
	})

	It("leaves non-TFX pipelines unchanged", func() {
		setAll()

		result, err := addHackedTFXPlatformSpec(pipeline, "kfpsdk")
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal(pipeline))
	})

	It("is idempotent", func() {
		setAll()

		once, err := addHackedTFXPlatformSpec(pipeline, "tfx")
		Expect(err).ToNot(HaveOccurred())
		twice, err := addHackedTFXPlatformSpec(once, "tfx")
		Expect(err).ToNot(HaveOccurred())
		Expect(twice).To(Equal(once))
	})

	DescribeTable("fails when a block is only partially configured",
		func(setting map[string]string) {
			for name, value := range setting {
				GinkgoT().Setenv(name, value)
			}

			_, err := addHackedTFXPlatformSpec(pipeline, "tfx")
			Expect(err).To(MatchError(ContainSubstring("must be set together")))
		},
		Entry("volume ConfigMap without mount path", map[string]string{volumeConfigMapEnvVar: volumeConfigMap}),
		Entry("mount path without volume ConfigMap", map[string]string{volumeMountPathEnvVar: volumeMountPath}),
		Entry("env ConfigMap without key", map[string]string{envConfigMapEnvVar: envConfigMap}),
		Entry("env key without ConfigMap", map[string]string{envKeyEnvVar: envKey}),
	)
})
