//go:build unit

package config

import (
	"testing"

	"github.com/kubeflow/pipelines/kubernetes_platform/go/kubernetesplatform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KFP Config Unit Suite")
}

var _ = Describe("Platform", func() {
	It("loads accelerators and Kubernetes settings from the parameter's environment variable", func() {
		GinkgoT().Setenv("PARAMETERS_PLATFORM", `{
			"accelerators": {"Nvidia-T4": {
				"resourceType": "nvidia.com/gpu",
				"nodeSelector": {"nvidia.com/gpu-type": "t4"},
				"tolerations": [{"key": "nvidia.com/gpu", "operator": "Exists", "effect": "NoSchedule"}]
			}},
			"secretAsVolume": [{"secretName": "cloud-credentials", "mountPath": "/var/secrets/cloud"}],
			"config_map_as_env": [{"configMapName": "cloud-env", "keyToEnv": [{"configMapKey": "CREDENTIALS", "envVar": "CREDENTIALS"}]}]
		}`)

		cfg, err := baseConfig.LoadConfig(Config{})

		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Parameters.Platform.Accelerators).To(Equal(Accelerators{"Nvidia-T4": {
			ResourceType: "nvidia.com/gpu",
			NodeSelector: map[string]string{"nvidia.com/gpu-type": "t4"},
			Tolerations:  []corev1.Toleration{{Key: "nvidia.com/gpu", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule}},
		}}))
		Expect(proto.Equal(cfg.Parameters.Platform.Kubernetes, &kubernetesplatform.KubernetesExecutorConfig{
			SecretAsVolume: []*kubernetesplatform.SecretAsVolume{{SecretName: "cloud-credentials", MountPath: "/var/secrets/cloud"}},
			ConfigMapAsEnv: []*kubernetesplatform.ConfigMapAsEnv{{
				ConfigMapName: "cloud-env",
				KeyToEnv:      []*kubernetesplatform.ConfigMapAsEnv_ConfigMapKeyToEnvMap{{ConfigMapKey: "CREDENTIALS", EnvVar: "CREDENTIALS"}},
			}},
		})).To(BeTrue())
	})

	DescribeTable("is empty",
		func(setEnv func()) {
			setEnv()

			cfg, err := baseConfig.LoadConfig(Config{})

			Expect(err).ToNot(HaveOccurred())
			Expect(cfg.Parameters.Platform.Accelerators).To(BeEmpty())
			Expect(proto.Size(cfg.Parameters.Platform.Kubernetes)).To(BeZero())
		},
		Entry("without the parameter", func() {}),
		Entry("for a null parameter", func() { GinkgoT().Setenv("PARAMETERS_PLATFORM", "null") }),
	)

	DescribeTable("rejects an invalid platform",
		func(platform string, expectedErr string) {
			GinkgoT().Setenv("PARAMETERS_PLATFORM", platform)

			_, err := baseConfig.LoadConfig(Config{})

			Expect(err).To(MatchError(ContainSubstring(expectedErr)))
		},
		Entry("accelerator without resourceType", `{"accelerators": {"nvidia-t4": {"nodeSelector": {"example.com/gpu-type": "t4"}}}}`, `accelerator "nvidia-t4" has no resourceType`),
		Entry("unknown accelerator field", `{"accelerators": {"nvidia-t4": {"resourceType": "nvidia.com/gpu", "nodeSelectors": {}}}}`, `unknown field "nodeSelectors"`),
		Entry("unknown setting", `{"secretAsVolumes": []}`, `unknown field "secretAsVolumes"`),
		Entry("invalid setting", `{"activeDeadlineSeconds": "soon"}`, "invalid platform"),
		Entry("invalid JSON", `{"accelerators": `, "invalid platform"),
	)
})
