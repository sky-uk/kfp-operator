//go:build unit

package provider

import (
	"encoding/json"

	"github.com/kubeflow/pipelines/kubernetes_platform/go/kubernetesplatform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
)

// kubernetesPlatformSpec returns a platform spec giving each executor the
// Kubernetes settings in executors.
func kubernetesPlatformSpec(executors map[string]any) string {
	GinkgoHelper()

	platformSpec, err := json.Marshal(map[string]any{
		"platforms": map[string]any{"kubernetes": map[string]any{"deploymentSpec": map[string]any{"executors": executors}}},
	})
	Expect(err).ToNot(HaveOccurred())
	return string(platformSpec)
}

var _ = Describe("Kubernetes executor settings", func() {
	poolSelector := map[string]any{"labels": map[string]any{"example.com/pool": "training"}}

	Context("KubernetesExecutorConfig", func() {
		It("reads settings under their JSON or proto names", func() {
			pipelinePackage := trainingPipeline(nil, kubernetesPlatformSpec(map[string]any{
				"exec-train": map[string]any{"nodeSelector": poolSelector, "active_deadline_seconds": "600"},
			}))

			settings, err := pipelinePackage.KubernetesExecutorConfig("exec-train")

			Expect(err).ToNot(HaveOccurred())
			Expect(settings.GetNodeSelector().GetLabels()).To(Equal(map[string]string{"example.com/pool": "training"}))
			Expect(settings.GetActiveDeadlineSeconds()).To(Equal(int64(600)))
		})

		It("returns empty settings for an executor without any", func() {
			settings, err := trainingPipeline(nil, "").KubernetesExecutorConfig("exec-train")

			Expect(err).ToNot(HaveOccurred())
			Expect(proto.Equal(settings, &kubernetesplatform.KubernetesExecutorConfig{})).To(BeTrue())
		})

		It("fails on settings that are not valid", func() {
			pipelinePackage := trainingPipeline(nil, kubernetesPlatformSpec(map[string]any{
				"exec-train": map[string]any{"activeDeadlineSeconds": "soon"},
			}))

			_, err := pipelinePackage.KubernetesExecutorConfig("exec-train")

			Expect(err).To(MatchError(ContainSubstring(`invalid Kubernetes settings for executor "exec-train"`)))
		})
	})

	Context("MergeKubernetesExecutorConfig", func() {
		It("adds settings to an executor without any, creating the platform spec", func() {
			pipelinePackage := trainingPipeline(nil, "")

			Expect(pipelinePackage.MergeKubernetesExecutorConfig("exec-train", &kubernetesplatform.KubernetesExecutorConfig{
				ImagePullSecret:       []*kubernetesplatform.ImagePullSecret{{SecretName: "registry"}},
				ActiveDeadlineSeconds: 600,
			})).To(Succeed())

			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"imagePullSecret":       []any{map[string]any{"secretName": "registry"}},
				"activeDeadlineSeconds": "600",
			}))
		})

		It("changes nothing for settings with no fields set", func() {
			pipelinePackage := trainingPipeline(nil, "")

			Expect(pipelinePackage.MergeKubernetesExecutorConfig("exec-train", &kubernetesplatform.KubernetesExecutorConfig{})).To(Succeed())

			Expect(pipelinePackage.PlatformSpec).To(BeNil())
		})

		It("keeps a field the executor already sets, under either name", func() {
			pipelinePackage := trainingPipeline(nil, kubernetesPlatformSpec(map[string]any{
				"exec-train": map[string]any{"node_selector": poolSelector, "activeDeadlineSeconds": "600"},
			}))

			Expect(pipelinePackage.MergeKubernetesExecutorConfig("exec-train", &kubernetesplatform.KubernetesExecutorConfig{
				NodeSelector:          &kubernetesplatform.NodeSelector{Labels: map[string]string{"example.com/gpu-type": "t4"}},
				ActiveDeadlineSeconds: 60,
			})).To(Succeed())

			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"node_selector":         poolSelector,
				"activeDeadlineSeconds": "600",
			}))
		})

		It("adds the list items an executor does not already have, under either name", func() {
			dedicated := map[string]any{"key": "dedicated", "operator": "Exists"}
			pipelinePackage := trainingPipeline(nil, kubernetesPlatformSpec(map[string]any{
				"exec-train": map[string]any{
					"tolerations":       []any{dedicated},
					"image_pull_secret": []any{map[string]any{"secretName": "registry"}},
				},
			}))

			Expect(pipelinePackage.MergeKubernetesExecutorConfig("exec-train", &kubernetesplatform.KubernetesExecutorConfig{
				Tolerations: []*kubernetesplatform.Toleration{
					{Key: "dedicated", Operator: "Exists"},
					{Key: "nvidia.com/gpu", Operator: "Exists", Effect: "NoSchedule"},
				},
				ImagePullSecret: []*kubernetesplatform.ImagePullSecret{{SecretName: "registry"}, {SecretName: "mirror"}},
			})).To(Succeed())

			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"tolerations": []any{
					dedicated,
					map[string]any{"key": "nvidia.com/gpu", "operator": "Exists", "effect": "NoSchedule"},
				},
				"image_pull_secret": []any{map[string]any{"secretName": "registry"}, map[string]any{"secretName": "mirror"}},
			}))
		})

		It("keeps settings this provider does not know and the settings of other executors", func() {
			pipelinePackage := trainingPipeline(nil, kubernetesPlatformSpec(map[string]any{
				"exec-train":    map[string]any{"futureSetting": "enabled"},
				"exec-evaluate": map[string]any{"activeDeadlineSeconds": "600"},
			}))

			Expect(pipelinePackage.MergeKubernetesExecutorConfig("exec-train", &kubernetesplatform.KubernetesExecutorConfig{
				ActiveDeadlineSeconds: 60,
			})).To(Succeed())

			Expect(kubernetesSettings(pipelinePackage, "exec-train")).To(Equal(map[string]any{
				"futureSetting":         "enabled",
				"activeDeadlineSeconds": "60",
			}))
			Expect(kubernetesSettings(pipelinePackage, "exec-evaluate")).To(Equal(map[string]any{"activeDeadlineSeconds": "600"}))
		})

		It("fails when a list field is not a list", func() {
			pipelinePackage := trainingPipeline(nil, kubernetesPlatformSpec(map[string]any{
				"exec-train": map[string]any{"image_pull_secret": "registry"},
			}))

			err := pipelinePackage.MergeKubernetesExecutorConfig("exec-train", &kubernetesplatform.KubernetesExecutorConfig{
				ImagePullSecret: []*kubernetesplatform.ImagePullSecret{{SecretName: "registry"}},
			})

			Expect(err).To(MatchError("image_pull_secret is not a list"))
		})
	})
})
