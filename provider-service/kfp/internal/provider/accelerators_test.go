//go:build unit

package provider

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	corev1 "k8s.io/api/core/v1"
)

// gpuPipeline returns a pipeline with an executor for each entry of
// executorResources, using the entry as the container's resources, followed by
// platformSpec when it is not empty.
func gpuPipeline(executorResources map[string]map[string]any, platformSpec string) *PipelinePackage {
	GinkgoHelper()

	executors := map[string]any{}
	for executor, resources := range executorResources {
		container := map[string]any{"image": "trainer"}
		if resources != nil {
			container["resources"] = resources
		}
		executors[executor] = map[string]any{"container": container}
	}
	pipelineSpec, err := json.Marshal(map[string]any{
		"deploymentSpec": map[string]any{"executors": executors},
		"pipelineInfo":   map[string]any{"name": "gpu"},
		"root":           map[string]any{},
		"schemaVersion":  "2.1.0",
	})
	Expect(err).ToNot(HaveOccurred())

	compiled := string(pipelineSpec)
	if platformSpec != "" {
		compiled += "\n---\n" + platformSpec
	}
	pipelinePackage, err := ParsePipelinePackage(resource.CompiledPipeline(compiled), "tfx")
	Expect(err).ToNot(HaveOccurred())
	return pipelinePackage
}

// acceleratorResources returns container resources requesting an accelerator,
// in the fields TFX writes.
func acceleratorResources(acceleratorType string, count string) map[string]any {
	return map[string]any{"accelerator": map[string]any{"resourceType": acceleratorType, "resourceCount": count}}
}

func acceleratorRequest(pipelinePackage *PipelinePackage, executor string) map[string]any {
	return nestedStruct(pipelinePackage.PipelineSpec.GetDeploymentSpec(), "executors", executor, "container", "resources", "accelerator").AsMap()
}

func kubernetesSettings(pipelinePackage *PipelinePackage, executor string) map[string]any {
	return pipelinePackage.kubernetesExecutorConfig(executor).AsMap()
}

var _ = Describe("AcceleratorConfig", func() {
	t4Toleration := map[string]any{"key": "nvidia.com/gpu", "operator": "Exists", "effect": "NoSchedule"}

	acceleratorConfig := &AcceleratorConfig{Accelerators: map[string]Accelerator{
		"nvidia-t4": {
			ResourceType: "nvidia.com/gpu",
			NodeSelector: map[string]string{"example.com/gpu-type": "t4"},
			Tolerations: []corev1.Toleration{{
				Key:      "nvidia.com/gpu",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			}},
		},
		"nvidia-l4": {
			ResourceType: "nvidia.com/gpu",
			NodeSelector: map[string]string{"example.com/gpu-type": "l4"},
		},
	}}

	Context("Apply", func() {
		It("rewrites the requested type and places the executor", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia-t4", "1")}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"resourceCount": "1",
				"resourceType":  "nvidia.com/gpu",
			}))
			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"nodeSelector": map[string]any{"labels": map[string]any{"example.com/gpu-type": "t4"}},
				"tolerations":  []any{t4Toleration},
			}))
		})

		It("rewrites the deprecated type that KFP SDK writes alongside resourceType", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": {"accelerator": map[string]any{
				"count":         "1",
				"resourceCount": "1",
				"resourceType":  "nvidia-t4",
				"type":          "nvidia-t4",
			}}}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"count":         "1",
				"resourceCount": "1",
				"resourceType":  "nvidia.com/gpu",
				"type":          "nvidia.com/gpu",
			}))
		})

		It("reads a request that only sets the deprecated type and count", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": {"accelerator": map[string]any{
				"count": 1,
				"type":  "nvidia-t4",
			}}}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"count": float64(1),
				"type":  "nvidia.com/gpu",
			}))
			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(HaveKey("nodeSelector"))
		})

		It("places each executor by its own accelerator type and leaves other executors alone", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{
				"exec-train":    acceleratorResources("nvidia-t4", "1"),
				"exec-evaluate": acceleratorResources("nvidia-l4", "2"),
				"exec-report":   {"cpuLimit": 2},
				"exec-publish":  nil,
			}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(HaveKeyWithValue(
				"nodeSelector", map[string]any{"labels": map[string]any{"example.com/gpu-type": "t4"}},
			))
			Expect(kubernetesSettings(pipelinePackage, "exec-evaluate")).To(Equal(map[string]any{
				"nodeSelector": map[string]any{"labels": map[string]any{"example.com/gpu-type": "l4"}},
			}))
			Expect(acceleratorRequest(pipelinePackage, "exec-evaluate")).To(HaveKeyWithValue("resourceType", "nvidia.com/gpu"))
			Expect(pipelinePackage.kubernetesExecutorConfig("exec-report")).To(BeNil())
			Expect(pipelinePackage.kubernetesExecutorConfig("exec-publish")).To(BeNil())
		})

		It("does not add a platform spec to a pipeline without accelerator requests", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-report": {"cpuLimit": 2}}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(pipelinePackage.PlatformSpec).To(BeNil())
		})

		DescribeTable("ignores requests the Kubeflow Pipelines driver does not act on",
			func(request map[string]any) {
				pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": {"accelerator": request}}, "")

				Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

				Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(HaveKeyWithValue("resourceType", "nvidia-t4"))
				Expect(pipelinePackage.PlatformSpec).To(BeNil())
			},
			Entry("no count", map[string]any{"resourceType": "nvidia-t4"}),
			Entry("a zero count", map[string]any{"resourceType": "nvidia-t4", "resourceCount": "0"}),
		)

		It("keeps a node selector the pipeline sets and adds the tolerations", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia-t4", "1")}, `platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec-train:
          nodeSelector:
            labels:
              example.com/pool: training
`)

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(HaveKeyWithValue("resourceType", "nvidia.com/gpu"))
			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"nodeSelector": map[string]any{"labels": map[string]any{"example.com/pool": "training"}},
				"tolerations":  []any{t4Toleration},
			}))
		})

		It("adds to the pipeline's tolerations without repeating one it already has", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia-t4", "1")}, `platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec-train:
          tolerations:
          - key: dedicated
            operator: Exists
          - effect: NoSchedule
            key: nvidia.com/gpu
            operator: Exists
`)

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(HaveKeyWithValue("tolerations", []any{
				map[string]any{"key": "dedicated", "operator": "Exists"},
				t4Toleration,
			}))
		})

		It("keeps other Kubernetes settings of the executor and of other executors", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia-t4", "1")}, `platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec-train:
          securityContext:
            runAsUser: '1000'
        exec-evaluate:
          activeDeadlineSeconds: '600'
`)

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(HaveKeyWithValue("securityContext", map[string]any{"runAsUser": "1000"}))
			Expect(kubernetesSettings(pipelinePackage, "exec-evaluate")).To(Equal(map[string]any{"activeDeadlineSeconds": "600"}))
		})

		It("writes toleration seconds", func() {
			tolerationSeconds := int64(300)
			config := &AcceleratorConfig{Accelerators: map[string]Accelerator{
				"nvidia-t4": {
					ResourceType: "nvidia.com/gpu",
					Tolerations: []corev1.Toleration{{
						Key:               "example.com/preemptible",
						Operator:          corev1.TolerationOpEqual,
						Value:             "true",
						Effect:            corev1.TaintEffectNoExecute,
						TolerationSeconds: &tolerationSeconds,
					}},
				},
			}}
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia-t4", "1")}, "")

			Expect(config.Apply(pipelinePackage)).To(Succeed())

			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"tolerations": []any{map[string]any{
					"key":               "example.com/preemptible",
					"operator":          "Equal",
					"value":             "true",
					"effect":            "NoExecute",
					"tolerationSeconds": "300",
				}},
			}))
		})

		It("only rewrites the type when the accelerator has no placement", func() {
			config := &AcceleratorConfig{Accelerators: map[string]Accelerator{"nvidia-t4": {ResourceType: "nvidia.com/gpu"}}}
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia-t4", "1")}, "")

			Expect(config.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(HaveKeyWithValue("resourceType", "nvidia.com/gpu"))
			Expect(pipelinePackage.PlatformSpec).To(BeNil())
		})

		It("fails on a type that is not configured, naming the executor and the configured types", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia-a100", "1")}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(MatchError(
				`executor "exec-train" requests accelerator type "nvidia-a100", which is not configured; configured types: nvidia-l4, nvidia-t4`,
			))
		})

		It("fails on a type that is only resolved at run time", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{
				"exec-train": acceleratorResources("{{$.inputs.parameters['pipelinechannel--gpu']}}", "1"),
			}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(MatchError(ContainSubstring("which is only resolved at run time")))
		})

		It("leaves a type that is not configured alone when the pipeline places the executor itself", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia.com/gpu", "1")}, `platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec-train:
          nodeSelector:
            labels:
              example.com/pool: training
`)

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(HaveKeyWithValue("resourceType", "nvidia.com/gpu"))
			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"nodeSelector": map[string]any{"labels": map[string]any{"example.com/pool": "training"}},
			}))
		})

		It("fails when the executor's tolerations are not a list", func() {
			pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia-t4", "1")}, `platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec-train:
          tolerations: dedicated
`)

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(MatchError(`failed to place executor "exec-train": tolerations is not a list`))
		})

		DescribeTable("changes nothing without accelerators configured",
			func(config *AcceleratorConfig) {
				pipelinePackage := gpuPipeline(map[string]map[string]any{"exec-train": acceleratorResources("nvidia-t4", "1")}, "")

				Expect(config.Apply(pipelinePackage)).To(Succeed())

				Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(HaveKeyWithValue("resourceType", "nvidia-t4"))
				Expect(pipelinePackage.PlatformSpec).To(BeNil())
			},
			Entry("nil", nil),
			Entry("empty", &AcceleratorConfig{}),
		)
	})

	Context("LoadAcceleratorConfig", func() {
		writeConfig := func(contents string) string {
			GinkgoHelper()
			path := filepath.Join(GinkgoT().TempDir(), "accelerators.yaml")
			Expect(os.WriteFile(path, []byte(contents), 0o600)).To(Succeed())
			return path
		}

		It("reads accelerators from YAML", func() {
			path := writeConfig(`accelerators:
  nvidia-t4:
    resourceType: nvidia.com/gpu
    nodeSelector:
      example.com/gpu-type: t4
    tolerations:
    - key: nvidia.com/gpu
      operator: Exists
      effect: NoSchedule
  nvidia-l4:
    resourceType: nvidia.com/gpu
`)

			loaded, err := LoadAcceleratorConfig(path)

			Expect(err).ToNot(HaveOccurred())
			Expect(loaded.Accelerators).To(Equal(map[string]Accelerator{
				"nvidia-t4": acceleratorConfig.Accelerators["nvidia-t4"],
				"nvidia-l4": {ResourceType: "nvidia.com/gpu"},
			}))
		})

		DescribeTable("rejects an invalid config",
			func(contents string, expectedErr string) {
				loaded, err := LoadAcceleratorConfig(writeConfig(contents))

				Expect(err).To(MatchError(ContainSubstring(expectedErr)))
				Expect(loaded).To(BeNil())
			},
			Entry("missing resourceType", "accelerators:\n  nvidia-t4:\n    nodeSelector:\n      example.com/gpu-type: t4\n", `accelerator "nvidia-t4" has no resourceType`),
			Entry("unknown field", "accelerators:\n  nvidia-t4:\n    resourceType: nvidia.com/gpu\n    nodeSelectors: {}\n", `unknown field "nodeSelectors"`),
			Entry("invalid YAML", "accelerators: [unclosed", "invalid accelerator config"),
		)

		It("fails when the file cannot be read", func() {
			_, err := LoadAcceleratorConfig(filepath.Join(GinkgoT().TempDir(), "missing.yaml"))

			Expect(err).To(MatchError(ContainSubstring("failed to read accelerator config")))
		})
	})
})
