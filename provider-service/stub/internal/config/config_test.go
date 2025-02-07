//go:build unit

package config

import (
	"net"
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	baseConfig "github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Unit Suite")
}

var _ = Context("load", func() {
	defaultConfig := Config{
		Server: baseConfig.Server{
			Host: net.IPv4zero.String(),
			Port: 8080,
		},
	}

	When("given no environment variable overrides", func() {
		It("correctly initializes the default config", func() {
			config, err := load()
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&defaultConfig))
		})
	})

	When("given an environment variable that matches a config key", func() {
		It("correctly overrides config value", func() {
			err := os.Setenv("SERVER_PORT", "8686")
			Expect(err).ToNot(HaveOccurred())

			expected := defaultConfig
			expected.Server.Port = 8686
			config, err := load()

			Expect(err).ToNot(HaveOccurred())
			Expect(config).To(Equal(&expected))
		})
	})
})
