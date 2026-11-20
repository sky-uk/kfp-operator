//go:build unit

package config

import (
	"context"
	"testing"

	"github.com/kubeflow/pipelines/kubernetes_platform/go/kubernetesplatform"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
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

	It("loads acceleratorsFrom", func() {
		GinkgoT().Setenv("PARAMETERS_PLATFORM", `{"acceleratorsFrom": {"configMapKeyRef": {"name": "accelerators", "key": "accelerators.yaml"}}}`)

		cfg, err := baseConfig.LoadConfig(Config{})

		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Parameters.Platform.AcceleratorsFrom).To(Equal(&AcceleratorsSource{ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
			LocalObjectReference: corev1.LocalObjectReference{Name: "accelerators"},
			Key:                  "accelerators.yaml",
		}}))
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
		Entry("acceleratorsFrom without a key", `{"acceleratorsFrom": {"configMapKeyRef": {"name": "accelerators"}}}`, "needs a configMapKeyRef with a name and a key"),
		Entry("unknown acceleratorsFrom field", `{"acceleratorsFrom": {"configMapRef": {"name": "accelerators"}}}`, `unknown field "configMapRef"`),
	)
})

var _ = Describe("ResolveAccelerators", func() {
	const namespace = "kfp-operator"
	t4 := Accelerator{ResourceType: "nvidia.com/gpu", NodeSelector: map[string]string{"nvidia.com/gpu-type": "t4"}}
	inlineT4 := Accelerator{ResourceType: "example.com/gpu"}

	configMap := func(data map[string]string) *corev1.ConfigMap {
		return &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "accelerators", Namespace: namespace}, Data: data}
	}
	platformFrom := func(key string, optional bool, inline Accelerators) Platform {
		ref := &corev1.ConfigMapKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "accelerators"}, Key: key, Optional: &optional}
		return Platform{Accelerators: inline, AcceleratorsFrom: &AcceleratorsSource{ConfigMapKeyRef: ref}}
	}
	resolve := func(platform *Platform, objects ...runtime.Object) error {
		scheme := runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())
		return platform.ResolveAccelerators(context.Background(), dynamicfake.NewSimpleDynamicClient(scheme, objects...), namespace)
	}
	acceleratorsYAML := map[string]string{"accelerators.yaml": "nvidia-t4:\n  resourceType: nvidia.com/gpu\n  nodeSelector:\n    nvidia.com/gpu-type: t4\nnvidia-l4:\n  resourceType: nvidia.com/gpu\n"}

	It("adds the ConfigMap's accelerators, which inline accelerators override", func() {
		platform := platformFrom("accelerators.yaml", false, Accelerators{"nvidia-l4": inlineT4})

		Expect(resolve(&platform, configMap(acceleratorsYAML))).To(Succeed())
		Expect(platform.Accelerators).To(Equal(Accelerators{"nvidia-t4": t4, "nvidia-l4": inlineT4}))
	})

	It("keeps the inline accelerators when an optional ConfigMap is missing", func() {
		platform := platformFrom("accelerators.yaml", true, Accelerators{"nvidia-l4": inlineT4})

		Expect(resolve(&platform)).To(Succeed())
		Expect(platform.Accelerators).To(Equal(Accelerators{"nvidia-l4": inlineT4}))
	})

	DescribeTable("fails",
		func(key string, expectedErr string, objects ...runtime.Object) {
			platform := platformFrom(key, false, nil)

			Expect(resolve(&platform, objects...)).To(MatchError(ContainSubstring(expectedErr)))
		},
		Entry("for a missing ConfigMap", "accelerators.yaml", `failed to get accelerators ConfigMap "accelerators"`),
		Entry("for a missing key", "other.yaml", `has no key "other.yaml"`, configMap(acceleratorsYAML)),
		Entry("for an accelerator without resourceType", "accelerators.yaml", `accelerator "nvidia-t4" has no resourceType`,
			configMap(map[string]string{"accelerators.yaml": "nvidia-t4:\n  nodeSelector: {}\n"})),
		Entry("for an unknown field", "accelerators.yaml", `unknown field "nodeSelectors"`,
			configMap(map[string]string{"accelerators.yaml": "nvidia-t4:\n  resourceType: nvidia.com/gpu\n  nodeSelectors: {}\n"})),
	)
})
