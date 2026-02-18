//go:build unit

package config

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Config Unit Suite")
}

var _ = Context("load", func() {
	defaultConfig := Config{
		Server: Server{
			Host: "0.0.0.0",
			Port: 8080,
		},
	}

	When("given no environment variable overrides", func() {
		It("correctly initialises the default config", func() {
			config, err := LoadConfig(defaultConfig)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&defaultConfig))
		})
	})
})
