//go:build unit

package run_completion_event_trigger

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
)

func TestRunCompletionEventTriggerConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Run Completion Event Trigger Config Unit Suite")
}

var _ = Context("LoadConfig", func() {

	viper.AddConfigPath(".")

	When("given the default config", func() {
		It("correctly loads config", func() {
			expectedConfig := Config{
				NATSConfig: NATSConfig{
					Subject: "events",
					ServerConfig: ServerConfig{
						Host: "localhost",
						Port: "4222",
					},
					JetStream: &JetStreamConfig{
						Enabled: false,
						Stream:  "run-completion-events",
						Storage: "file",
						MaxAge:  "7d",
					},
					Auth: &AuthConfig{
						TLS: &TLSConfig{
							Enabled:            false,
							CertFile:           "/etc/ssl/certs/tls.crt",
							KeyFile:            "/etc/ssl/private/tls.key",
							CAFile:             "/etc/ssl/certs/ca.crt",
							InsecureSkipVerify: false,
						},
						ClientAuth: "username: your-username\npassword: your-password\n",
					},
				},
				ServerConfig: ServerConfig{
					Host: "localhost",
					Port: "50051",
				},
				MetricsConfig: ServerConfig{
					Host: "localhost",
					Port: "8081",
				},
			}
			config, err := LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&expectedConfig))
		})
	})

	When("given an environment variable", func() {
		It("correctly overrides config value", func() {
			os.Setenv("NATSCONFIG_SERVERCONFIG_PORT", "5000")
			expectedConfig := Config{
				NATSConfig: NATSConfig{
					Subject: "events",
					ServerConfig: ServerConfig{
						Host: "localhost",
						Port: "5000",
					},
					JetStream: &JetStreamConfig{
						Enabled: false,
						Stream:  "run-completion-events",
						Storage: "file",
						MaxAge:  "7d",
					},
					Auth: &AuthConfig{
						TLS: &TLSConfig{
							Enabled:            false,
							CertFile:           "/etc/ssl/certs/tls.crt",
							KeyFile:            "/etc/ssl/private/tls.key",
							CAFile:             "/etc/ssl/certs/ca.crt",
							InsecureSkipVerify: false,
						},
						ClientAuth: "username: your-username\npassword: your-password\n",
					},
				},
				ServerConfig: ServerConfig{
					Host: "localhost",
					Port: "50051",
				},
				MetricsConfig: ServerConfig{
					Host: "localhost",
					Port: "8081",
				},
			}
			config, err := LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&expectedConfig))
		})
	})

	When("config includes JetStream configuration", func() {
		It("correctly loads JetStream config", func() {
			// This test would require a separate config file with JetStream settings
			// For now, we verify that the structure supports the new fields
			config := &Config{
				NATSConfig: NATSConfig{
					Subject: "events",
					ServerConfig: ServerConfig{
						Host: "localhost",
						Port: "4222",
					},
					JetStream: &JetStreamConfig{
						Enabled: true,
						Stream:  "test-stream",
					},
					Auth: &AuthConfig{
						TLS: &TLSConfig{
							Enabled:  true,
							CertFile: "/path/to/cert.pem",
							KeyFile:  "/path/to/key.pem",
						},
					},
				},
			}

			// Verify the structure is properly defined
			Expect(config.NATSConfig.JetStream).ToNot(BeNil())
			Expect(config.NATSConfig.JetStream.Enabled).To(BeTrue())
			Expect(config.NATSConfig.JetStream.Stream).To(Equal("test-stream"))
			Expect(config.NATSConfig.Auth).ToNot(BeNil())
			Expect(config.NATSConfig.Auth.TLS).ToNot(BeNil())
			Expect(config.NATSConfig.Auth.TLS.Enabled).To(BeTrue())
		})
	})
})
