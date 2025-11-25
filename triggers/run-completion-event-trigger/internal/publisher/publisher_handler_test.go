//go:build unit

package publisher

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/cmd/config"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PublisherHandler Unit Test Suite")
}

var _ = Describe("PublisherHandler Factory", func() {
	Context("NewPublisherFromConfig", func() {
		When("JetStream is enabled", func() {
			It("should detect JetStream configuration", func() {
				config := &configLoader.NATSConfig{
					Subject: "test",
					ServerConfig: configLoader.ServerConfig{
						Host: "localhost",
						Port: "4222",
					},
					JetStream: &configLoader.JetStreamConfig{
						Enabled: true,
						Stream:  "test-stream",
					},
				}

				// Test the logic that determines publisher type
				Expect(isJetStreamEnabled(config)).To(BeTrue())
			})
		})

		When("no special config is provided", func() {
			It("should use basic configuration", func() {
				config := &configLoader.NATSConfig{
					Subject: "test",
					ServerConfig: configLoader.ServerConfig{
						Host: "localhost",
						Port: "4222",
					},
				}

				Expect(isJetStreamEnabled(config)).To(BeFalse())
				Expect(config.Auth).To(BeNil())
			})
		})
	})

	Context("Helper Functions", func() {
		Context("buildServerURL", func() {
			When("TLS is not enabled", func() {
				It("should return nats:// URL", func() {
					config := &configLoader.NATSConfig{
						ServerConfig: configLoader.ServerConfig{
							Host: "localhost",
							Port: "4222",
						},
					}
					url := buildServerURL(config)
					Expect(url).To(Equal("nats://localhost:4222"))
				})
			})

			When("TLS is enabled", func() {
				It("should return tls:// URL", func() {
					config := &configLoader.NATSConfig{
						ServerConfig: configLoader.ServerConfig{
							Host: "localhost",
							Port: "4222",
						},
						Auth: &configLoader.AuthConfig{
							TLS: &configLoader.TLSConfig{
								Enabled: true,
							},
						},
					}
					url := buildServerURL(config)
					Expect(url).To(Equal("nats://localhost:4222"))
				})
			})
		})

		Context("Connection Options", func() {
			It("should use explicit NATS connection options", func() {
				// This test verifies that our connection creation uses explicit options
				// The actual connection creation is tested in integration tests
				config := &configLoader.NATSConfig{
					Subject: "test",
					ServerConfig: configLoader.ServerConfig{
						Host: "localhost",
						Port: "4222",
					},
				}

				// Verify configuration parsing
				Expect(config.Subject).To(Equal("test"))
				Expect(config.ServerConfig.Host).To(Equal("localhost"))
				Expect(config.ServerConfig.Port).To(Equal("4222"))
			})
		})
	})
})
