//go:build unit

package nats_event_trigger

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var validConfig = `
natsConfig:
  subject: aSubject
  url: https://example.com
serverConfig:
  host: https://example.com
  port: 50051
`
var invalidYaml = `foo: "bar`
var validYamlMissingFields = `foo: "bar"`
var invalidConfig = `natsConfig: events`

var _ = Context("LoadConfig", func() {

	When("given a valid filename", func() {
		It("returns returns a valid Config object", func() {
			reader := strings.NewReader(validConfig)

			subject := "aSubject"
			url := "https://example.com"
			host := "https://example.com"
			port := "50051"

			expectedConfig := &Config{
				NATSConfig: &NATSConfig{
					Subject: &subject,
					Url:     &url,
				},
				ServerConfig: &ServerConfig{
					Host: &host,
					Port: &port,
				},
			}
			config, err := LoadConfig(reader)
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(expectedConfig))
		})
	})

	When("given invalid yaml", func() {
		It("fails to create config", func() {
			reader := strings.NewReader(invalidYaml)
			config, err := LoadConfig(reader)
			Expect(config).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	When("given invalid config", func() {
		It("fails to create config", func() {
			reader := strings.NewReader(invalidConfig)
			config, err := LoadConfig(reader)
			Expect(config).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})

	When("given config with missing fields", func() {
		It("fails to create config", func() {
			reader := strings.NewReader(validYamlMissingFields)
			config, err := LoadConfig(reader)
			Expect(config).To(BeNil())
			Expect(err).To(HaveOccurred())
		})
	})
})
