//go:build unit

package config

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Unit Suite")
}

var _ = Context("load", func() {
	viper.AddConfigPath(".")

	expectedConfig := Config{
		ProviderName:    "local",
		OperatorWebhook: "http://localhost:8080/events",
		Pod: Pod{
			Namespace: "provider-namespace",
		},
	}

	When("given the default config", func() {
		It("correctly loads config", func() {
			config, err := load()
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&expectedConfig))
		})
	})

	When("given a pod namespace environment variable", func() {
		It("correctly overrides config value", func() {
			err := os.Setenv("POD_NAMESPACE", "kfp-operator-system")
			Expect(err).NotTo(HaveOccurred())
			expectedConfig.Pod.Namespace = "kfp-operator-system"
			config, err := load()
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&expectedConfig))
		})
	})
})
