//go:build unit

package config

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	corev1 "k8s.io/api/core/v1"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "KFP Config Unit Suite")
}

var _ = Describe("Accelerators", func() {
	It("loads from the parameter's environment variable, keeping its keys as written", func() {
		GinkgoT().Setenv("PARAMETERS_ACCELERATORS", `{"Nvidia-T4": {
			"resourceType": "nvidia.com/gpu",
			"nodeSelector": {"nvidia.com/gpu-type": "t4"},
			"tolerations": [{"key": "nvidia.com/gpu", "operator": "Exists", "effect": "NoSchedule"}]
		}}`)

		cfg, err := baseConfig.LoadConfig(Config{})

		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Parameters.Accelerators).To(Equal(Accelerators{"Nvidia-T4": {
			ResourceType: "nvidia.com/gpu",
			NodeSelector: map[string]string{"nvidia.com/gpu-type": "t4"},
			Tolerations:  []corev1.Toleration{{Key: "nvidia.com/gpu", Operator: corev1.TolerationOpExists, Effect: corev1.TaintEffectNoSchedule}},
		}}))
	})

	It("is empty without the parameter", func() {
		cfg, err := baseConfig.LoadConfig(Config{})

		Expect(err).ToNot(HaveOccurred())
		Expect(cfg.Parameters.Accelerators).To(BeEmpty())
	})

	DescribeTable("rejects invalid accelerators",
		func(accelerators string, expectedErr string) {
			GinkgoT().Setenv("PARAMETERS_ACCELERATORS", accelerators)

			_, err := baseConfig.LoadConfig(Config{})

			Expect(err).To(MatchError(ContainSubstring(expectedErr)))
		},
		Entry("missing resourceType", `{"nvidia-t4": {"nodeSelector": {"example.com/gpu-type": "t4"}}}`, `"nvidia-t4" has no resourceType`),
		Entry("unknown field", `{"nvidia-t4": {"resourceType": "nvidia.com/gpu", "nodeSelectors": {}}}`, `unknown field "nodeSelectors"`),
		Entry("invalid JSON", `{"nvidia-t4": `, "invalid accelerators"),
	)
})
