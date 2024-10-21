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
				},
				ServerConfig: ServerConfig{
					Host: "localhost",
					Port: "50051",
				},
			}
			config, err := LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&expectedConfig))
		})
	})

	When("given an enviornment variable", func() {
		It("correctly override config value", func() {
			os.Setenv("NATSCONFIG_SERVERCONFIG_PORT", "5000")
			expectedConfig := Config{
				NATSConfig: NATSConfig{
					Subject: "events",
					ServerConfig: ServerConfig{
						Host: "localhost",
						Port: "5000",
					},
				},
				ServerConfig: ServerConfig{
					Host: "localhost",
					Port: "50051",
				},
			}
			config, err := LoadConfig()
			Expect(err).NotTo(HaveOccurred())
			Expect(config).To(Equal(&expectedConfig))
		})
	})
})
