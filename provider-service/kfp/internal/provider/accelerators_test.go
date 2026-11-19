//go:build unit

package provider

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	corev1 "k8s.io/api/core/v1"
)

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

func trainingPipeline(resources map[string]any, platformSpec string) *PipelinePackage {
	GinkgoHelper()
	return gpuPipeline(map[string]map[string]any{"exec-train": resources}, platformSpec)
}

func acceleratorResources(acceleratorType string, count string) map[string]any {
	return map[string]any{"accelerator": map[string]any{"resourceType": acceleratorType, "resourceCount": count}}
}

func acceleratorRequest(pipelinePackage *PipelinePackage, executor string) map[string]any {
	return nestedStruct(pipelinePackage.PipelineSpec.GetDeploymentSpec(), "executors", executor, "container", "resources", "accelerator").AsMap()
}

func kubernetesSettings(pipelinePackage *PipelinePackage, executor string) map[string]any {
	return pipelinePackage.kubernetesExecutorStruct(executor).AsMap()
}

var _ = Describe("placeAccelerators", func() {
	t4Selector := map[string]any{"labels": map[string]any{"example.com/gpu-type": "t4"}}
	t4Toleration := map[string]any{"key": "nvidia.com/gpu", "operator": "Exists", "effect": "NoSchedule"}
	poolSelector := map[string]any{"labels": map[string]any{"example.com/pool": "training"}}
	tolerationSeconds := int64(300)

	accelerators := config.Accelerators{
		"nvidia-t4": {
			ResourceType: "nvidia.com/gpu",
			NodeSelector: map[string]string{"example.com/gpu-type": "t4"},
			Tolerations:  []corev1.Toleration{{Key: "nvidia.com/gpu", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule}},
		},
		"nvidia-l4": {
			ResourceType: "nvidia.com/gpu",
			Tolerations: []corev1.Toleration{{
				Key:               "example.com/preemptible",
				Operator:          corev1.TolerationOpEqual,
				Value:             "true",
				Effect:            corev1.TaintEffectNoExecute,
				TolerationSeconds: &tolerationSeconds,
			}},
		},
	}

	DescribeTable("rewrites the requested type and places the executor",
		func(request map[string]any, expected map[string]any) {
			pipelinePackage := trainingPipeline(map[string]any{"accelerator": request}, "")

			Expect(placeAccelerators(accelerators, pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(Equal(expected))
			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"nodeSelector": t4Selector,
				"tolerations":  []any{t4Toleration},
			}))
		},
		Entry("resourceType",
			map[string]any{"resourceType": "nvidia-t4", "resourceCount": "1"},
			map[string]any{"resourceType": "nvidia.com/gpu", "resourceCount": "1"}),
		Entry("resourceType and the deprecated type",
			map[string]any{"resourceType": "nvidia-t4", "resourceCount": "1", "type": "nvidia-t4", "count": "1"},
			map[string]any{"resourceType": "nvidia.com/gpu", "resourceCount": "1", "type": "nvidia.com/gpu", "count": "1"}),
		Entry("only the deprecated type",
			map[string]any{"type": "nvidia-t4", "count": 1},
			map[string]any{"type": "nvidia.com/gpu", "count": float64(1)}),
	)

	It("places each executor by its own type and leaves executors without a request alone", func() {
		pipelinePackage := gpuPipeline(map[string]map[string]any{
			"exec-train":    acceleratorResources("nvidia-t4", "1"),
			"exec-evaluate": acceleratorResources("nvidia-l4", "2"),
			"exec-report":   {"cpuLimit": 2},
			"exec-publish":  nil,
		}, "")

		Expect(placeAccelerators(accelerators, pipelinePackage)).To(Succeed())

		Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(HaveKeyWithValue("nodeSelector", t4Selector))
		Expect(kubernetesSettings(pipelinePackage, "exec-evaluate")).To(Equal(map[string]any{
			"tolerations": []any{map[string]any{
				"key":               "example.com/preemptible",
				"operator":          "Equal",
				"value":             "true",
				"effect":            "NoExecute",
				"tolerationSeconds": "300",
			}},
		}))
		Expect(pipelinePackage.kubernetesExecutorStruct("exec-report")).To(BeNil())
		Expect(pipelinePackage.kubernetesExecutorStruct("exec-publish")).To(BeNil())
	})

	DescribeTable("ignores requests the Kubeflow Pipelines driver does not act on",
		func(request map[string]any) {
			pipelinePackage := trainingPipeline(map[string]any{"accelerator": request}, "")

			Expect(placeAccelerators(accelerators, pipelinePackage)).To(Succeed())

			Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(Equal(request))
			Expect(pipelinePackage.PlatformSpec).To(BeNil())
		},
		Entry("no count", map[string]any{"resourceType": "nvidia-t4"}),
		Entry("a zero count", map[string]any{"resourceType": "nvidia-t4", "resourceCount": "0"}),
	)

	It("keeps a node selector the pipeline sets and adds the tolerations", func() {
		pipelinePackage := trainingPipeline(acceleratorResources("nvidia-t4", "1"), kubernetesPlatformSpec(map[string]any{
			"exec-train": map[string]any{"nodeSelector": poolSelector},
		}))

		Expect(placeAccelerators(accelerators, pipelinePackage)).To(Succeed())

		Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(HaveKeyWithValue("resourceType", "nvidia.com/gpu"))
		Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
			"nodeSelector": poolSelector,
			"tolerations":  []any{t4Toleration},
		}))
	})

	It("leaves a type that is not configured alone when the pipeline sets a node selector", func() {
		pipelinePackage := trainingPipeline(acceleratorResources("nvidia.com/gpu", "1"), kubernetesPlatformSpec(map[string]any{
			"exec-train": map[string]any{"nodeSelector": poolSelector},
		}))

		Expect(placeAccelerators(accelerators, pipelinePackage)).To(Succeed())

		Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(HaveKeyWithValue("resourceType", "nvidia.com/gpu"))
		Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{"nodeSelector": poolSelector}))
	})

	DescribeTable("fails on a request it cannot place",
		func(request map[string]any, expectedErr string) {
			pipelinePackage := trainingPipeline(map[string]any{"accelerator": request}, "")

			Expect(placeAccelerators(accelerators, pipelinePackage)).To(MatchError(ContainSubstring(expectedErr)))
		},
		Entry("a type that is not configured",
			map[string]any{"resourceType": "nvidia-a100", "resourceCount": "1"},
			`executor "exec-train" requests accelerator type "nvidia-a100", which is not configured; configured types: nvidia-l4, nvidia-t4`),
		Entry("a type resolved at run time",
			map[string]any{"resourceType": "{{$.inputs.parameters['pipelinechannel--gpu']}}", "resourceCount": "1"},
			"which is only resolved at run time"),
		Entry("an invalid request",
			map[string]any{"type": "nvidia-t4", "count": "one"},
			`executor "exec-train" has an invalid accelerator request`),
	)

	It("changes nothing without accelerators configured", func() {
		pipelinePackage := trainingPipeline(acceleratorResources("nvidia-t4", "1"), "")

		Expect(placeAccelerators(nil, pipelinePackage)).To(Succeed())

		Expect(acceleratorRequest(pipelinePackage, "exec-train")).To(HaveKeyWithValue("resourceType", "nvidia-t4"))
		Expect(pipelinePackage.PlatformSpec).To(BeNil())
	})
})
