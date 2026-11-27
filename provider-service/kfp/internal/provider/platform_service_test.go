//go:build unit

package provider

import (
	"github.com/kubeflow/pipelines/kubernetes_platform/go/kubernetesplatform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
)

var _ = Describe("DefaultPlatformService", func() {
	gpuPipelineYAML := resource.CompiledPipeline(`deploymentSpec:
  executors:
    exec-train:
      container:
        image: trainer
        resources:
          accelerator:
            resourceCount: '1'
            resourceType: nvidia-t4
pipelineInfo:
  name: test
root: {}
schemaVersion: 2.1.0
`)
	invalidPipelineYAML := resource.CompiledPipeline("unknownField: true\n")

	t4 := config.Accelerators{"nvidia-t4": {ResourceType: "nvidia.com/gpu"}}

	Context("Apply", func() {
		It("should return the pipeline untouched without platform settings", func() {
			result, err := NewDefaultPlatformService(config.Platform{}).Apply(invalidPipelineYAML)

			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(invalidPipelineYAML))
		})

		It("should place accelerators", func() {
			result, err := NewDefaultPlatformService(config.Platform{Accelerators: t4}).Apply(gpuPipelineYAML)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(ContainSubstring("resourceType: nvidia.com/gpu"))
		})

		It("should apply platform settings without accelerators", func() {
			platform := config.Platform{
				Kubernetes: &kubernetesplatform.KubernetesExecutorConfig{
					SecretAsVolume: []*kubernetesplatform.SecretAsVolume{{SecretName: "cloud-credentials", MountPath: "/var/secrets/cloud"}},
				},
			}

			result, err := NewDefaultPlatformService(platform).Apply(gpuPipelineYAML)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(ContainSubstring("secretName: cloud-credentials"))
		})

		It("should place accelerators before applying the platform settings", func() {
			platform := config.Platform{
				Accelerators: config.Accelerators{"nvidia-t4": {ResourceType: "nvidia.com/gpu", NodeSelector: map[string]string{"example.com/gpu-type": "t4"}}},
				Kubernetes: &kubernetesplatform.KubernetesExecutorConfig{
					NodeSelector:   &kubernetesplatform.NodeSelector{Labels: map[string]string{"example.com/pool": "general"}},
					SecretAsVolume: []*kubernetesplatform.SecretAsVolume{{SecretName: "cloud-credentials", MountPath: "/var/secrets/cloud"}},
				},
			}

			result, err := NewDefaultPlatformService(platform).Apply(gpuPipelineYAML)

			Expect(err).ToNot(HaveOccurred())
			Expect(string(result)).To(ContainSubstring("example.com/gpu-type: t4"))
			Expect(string(result)).ToNot(ContainSubstring("example.com/pool"))
			Expect(string(result)).To(ContainSubstring("secretName: cloud-credentials"))
		})

		It("should return err if the compiled pipeline is invalid", func() {
			result, err := NewDefaultPlatformService(config.Platform{Accelerators: t4}).Apply(invalidPipelineYAML)

			Expect(err).To(MatchError(ContainSubstring("failed to parse compiled pipeline")))
			Expect(result).To(BeNil())
		})

		It("should return err if an accelerator type is not configured", func() {
			platform := config.Platform{
				Accelerators: config.Accelerators{"nvidia-l4": {ResourceType: "nvidia.com/gpu"}},
				Kubernetes:   &kubernetesplatform.KubernetesExecutorConfig{NodeSelector: &kubernetesplatform.NodeSelector{Labels: map[string]string{"example.com/pool": "general"}}},
			}

			result, err := NewDefaultPlatformService(platform).Apply(gpuPipelineYAML)

			Expect(err).To(MatchError(ContainSubstring("failed to place accelerators")))
			Expect(result).To(BeNil())
		})
	})
})
