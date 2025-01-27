//go:build unit

package config

import (
	"context"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Unit Suite")
}

var _ = Context("load", func() {
	ctx := context.Background()
	expectedConfig := Config{}

	When("given no environment variable overrides", func() {
		It("correctly initialises empty config", func() {
			config, err := LoadConfig(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&expectedConfig))
		})
	})

	When("given a pod namespace environment variable", func() {
		It("correctly overrides config value", func() {
			err := os.Setenv("POD_NAMESPACE", "kfp-operator-system")
			Expect(err).NotTo(HaveOccurred())
			expectedConfig.Pod.Namespace = "kfp-operator-system"
			config, err := LoadConfig(ctx)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&expectedConfig))
		})
	})
})
