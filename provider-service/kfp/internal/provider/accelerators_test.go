//go:build unit

package provider

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	corev1 "k8s.io/api/core/v1"
)

// gpuPipeline returns a pipeline whose executors have the given container
// resources, followed by platformSpec if it is not empty.
func gpuPipeline(executorResources map[string]string, platformSpec string) *PipelinePackage {
	compiled := "deploymentSpec:\n  executors:\n"
	for executor, resources := range executorResources {
		compiled += "    " + executor + ":\n      container:\n        image: trainer\n"
		if resources != "" {
			compiled += "        resources:\n" + resources
		}
	}
	compiled += "pipelineInfo:\n  name: gpu\nroot: {}\nschemaVersion: 2.1.0\n"
	if platformSpec != "" {
		compiled += "---\n" + platformSpec
	}

	pipelinePackage, err := ParsePipelinePackage(resource.CompiledPipeline(compiled), "tfx")
	Expect(err).ToNot(HaveOccurred())
	return pipelinePackage
}

func acceleratorRequest(pipelinePackage *PipelinePackage, executor string) map[string]any {
	return nestedStruct(pipelinePackage.PipelineSpec.GetDeploymentSpec(), "executors", executor, "container", "resources", "accelerator").AsMap()
}

func kubernetesExecutorConfig(pipelinePackage *PipelinePackage, executor string) map[string]any {
	return pipelinePackage.KubernetesExecutorConfig(executor).AsMap()
}

const (
	t4Request = `          accelerator:
            resourceCount: '1'
            resourceType: nvidia-t4
`
	sdkT4Request = `          accelerator:
            count: '1'
            resourceCount: '1'
            resourceType: nvidia-t4
            type: nvidia-t4
`
	l4Request = `          accelerator:
            resourceCount: '2'
            resourceType: nvidia-l4
`
	cpuRequest = `          cpuLimit: 2.0
`
)

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
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": t4Request}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"resourceCount": "1",
				"resourceType":  "nvidia.com/gpu",
			}))
			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"nodeSelector": map[string]any{"labels": map[string]any{"example.com/gpu-type": "t4"}},
				"tolerations":  []any{t4Toleration},
			}))
		})

		It("rewrites the deprecated type field as well when the request sets it", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": sdkT4Request}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"count":         "1",
				"resourceCount": "1",
				"resourceType":  "nvidia.com/gpu",
				"type":          "nvidia.com/gpu",
			}))
		})

		It("reads a request that only sets the deprecated fields", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": "          accelerator:\n            count: 1\n            type: nvidia-t4\n"}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"count": float64(1),
				"type":  "nvidia.com/gpu",
			}))
			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-train")).To(HaveKey("nodeSelector"))
		})

		It("places each executor by its own accelerator type and leaves other executors alone", func() {
			pipelinePackage := gpuPipeline(map[string]string{
				"exec-train":    t4Request,
				"exec-evaluate": l4Request,
				"exec-report":   cpuRequest,
				"exec-publish":  "",
			}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-train")["nodeSelector"]).To(Equal(
				map[string]any{"labels": map[string]any{"example.com/gpu-type": "t4"}},
			))
			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-evaluate")).To(Equal(map[string]any{
				"nodeSelector": map[string]any{"labels": map[string]any{"example.com/gpu-type": "l4"}},
			}))
			Expect(acceleratorRequest(pipelinePackage, "exec-evaluate")["resourceType"]).To(Equal("nvidia.com/gpu"))
			Expect(pipelinePackage.KubernetesExecutorConfig("exec-report")).To(BeNil())
			Expect(pipelinePackage.KubernetesExecutorConfig("exec-publish")).To(BeNil())
		})

		It("does not add a platform spec to a pipeline without accelerator requests", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-report": cpuRequest}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(pipelinePackage.PlatformSpec).To(BeNil())
		})

		DescribeTable("ignores requests the Kubeflow Pipelines driver would not act on",
			func(request string) {
				pipelinePackage := gpuPipeline(map[string]string{"exec-train": request}, "")

				Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

				Expect(acceleratorRequest(pipelinePackage, "exec-train")["resourceType"]).To(Equal("nvidia-t4"))
				Expect(pipelinePackage.PlatformSpec).To(BeNil())
			},
			Entry("no count", "          accelerator:\n            resourceType: nvidia-t4\n"),
			Entry("a zero count", "          accelerator:\n            resourceCount: '0'\n            resourceType: nvidia-t4\n"),
		)

		It("keeps a node selector the pipeline sets and adds the tolerations", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": t4Request}, `platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec-train:
          nodeSelector:
            labels:
              example.com/pool: training
`)

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")["resourceType"]).To(Equal("nvidia.com/gpu"))
			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"nodeSelector": map[string]any{"labels": map[string]any{"example.com/pool": "training"}},
				"tolerations":  []any{t4Toleration},
			}))
		})

		It("adds to the pipeline's tolerations without repeating one it already has", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": t4Request}, `platforms:
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

			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-train")["tolerations"]).To(Equal([]any{
				map[string]any{"key": "dedicated", "operator": "Exists"},
				t4Toleration,
			}))
		})

		It("keeps other Kubernetes settings of the executor and of other executors", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": t4Request}, `platforms:
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

			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-train")).To(HaveKeyWithValue("securityContext", map[string]any{"runAsUser": "1000"}))
			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-evaluate")).To(Equal(map[string]any{"activeDeadlineSeconds": "600"}))
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
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": t4Request}, "")

			Expect(config.Apply(pipelinePackage)).To(Succeed())

			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-train")).To(Equal(map[string]any{
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
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": t4Request}, "")

			Expect(config.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")["resourceType"]).To(Equal("nvidia.com/gpu"))
			Expect(pipelinePackage.PlatformSpec).To(BeNil())
		})

		It("fails on a type that is not configured, naming the executor and the configured types", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": "          accelerator:\n            resourceCount: '1'\n            resourceType: nvidia-a100\n"}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(MatchError(
				`executor "exec-train" requests accelerator type "nvidia-a100", which is not configured; configured types: nvidia-l4, nvidia-t4`,
			))
		})

		It("fails on a type that is only resolved at run time", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": "          accelerator:\n            resourceCount: '1'\n            resourceType: \"{{$.inputs.parameters['pipelinechannel--gpu']}}\"\n"}, "")

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(MatchError(ContainSubstring("which is only resolved at run time")))
		})

		It("leaves a type that is not configured alone when the pipeline places the executor itself", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": "          accelerator:\n            resourceCount: '1'\n            resourceType: nvidia.com/gpu\n"}, `platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec-train:
          nodeSelector:
            labels:
              example.com/pool: training
`)

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")["resourceType"]).To(Equal("nvidia.com/gpu"))
			Expect(kubernetesExecutorConfig(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"nodeSelector": map[string]any{"labels": map[string]any{"example.com/pool": "training"}},
			}))
		})

		It("fails when the executor's tolerations are not a list", func() {
			pipelinePackage := gpuPipeline(map[string]string{"exec-train": t4Request}, `platforms:
  kubernetes:
    deploymentSpec:
      executors:
        exec-train:
          tolerations: dedicated
`)

			Expect(acceleratorConfig.Apply(pipelinePackage)).To(MatchError(ContainSubstring(`failed to place executor "exec-train": tolerations is not a list`)))
		})

		DescribeTable("changes nothing without accelerators configured",
			func(config *AcceleratorConfig) {
				pipelinePackage := gpuPipeline(map[string]string{"exec-train": t4Request}, "")

				Expect(config.Apply(pipelinePackage)).To(Succeed())

				Expect(acceleratorRequest(pipelinePackage, "exec-train")["resourceType"]).To(Equal("nvidia-t4"))
				Expect(pipelinePackage.PlatformSpec).To(BeNil())
			},
			Entry("nil", nil),
			Entry("empty", &AcceleratorConfig{}),
		)
	})

	Context("LoadAcceleratorConfig", func() {
		writeConfig := func(contents string) string {
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
