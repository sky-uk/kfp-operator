//go:build unit

package publisher

import (
	"errors"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	configLoader "github.com/sky-uk/kfp-operator/triggers/run-completion-event-trigger/cmd/config"
)

// Mock NATS connection for JetStream tests
type mockNatsConnJS struct {
	isConnectedFunc func() bool
}

func (m *mockNatsConnJS) Publish(subject string, data []byte) error {
	return nil
}

func (m *mockNatsConnJS) IsConnected() bool {
	if m.isConnectedFunc != nil {
		return m.isConnectedFunc()
	}
	return true
}

var _ = Describe("JetStreamPublisher", func() {
	var (
		publisher *JetStreamPublisher
		mockConn  *mockNatsConnJS
		config    *configLoader.NATSConfig
	)

	BeforeEach(func() {
		mockConn = &mockNatsConnJS{}

		config = &configLoader.NATSConfig{
			Subject: "test-events",
			ServerConfig: configLoader.ServerConfig{
				Host: "localhost",
				Port: "4222",
			},
			JetStream: &configLoader.JetStreamConfig{
				Enabled: true,
				Stream:  "test-stream",
				Storage: "file",
				MaxAge:  "24h",
			},
		}

		publisher = &JetStreamPublisher{
			NatsConn:  mockConn,
			Subject:   config.Subject,
			JetStream: nil, // We'll test methods that don't require JetStream
			Config:    config,
		}
	})

	Context("Name", func() {
		It("should return the correct publisher name", func() {
			Expect(publisher.Name()).To(Equal("nats-jetstream-publisher"))
		})
	})

	Context("IsHealthy", func() {
		When("NATS connection is healthy", func() {
			It("should return true", func() {
				mockConn.isConnectedFunc = func() bool { return true }
				Expect(publisher.IsHealthy()).To(BeTrue())
			})
		})

		When("NATS connection is unhealthy", func() {
			It("should return false", func() {
				mockConn.isConnectedFunc = func() bool { return false }
				Expect(publisher.IsHealthy()).To(BeFalse())
			})
		})
	})

	Context("createStreamConfig", func() {
		When("using default configuration", func() {
			It("should create config with file storage and 24h retention", func() {
				config, err := publisher.createStreamConfig()

				Expect(err).ToNot(HaveOccurred())
				Expect(config.Name).To(Equal("test-stream"))
				Expect(config.Subjects).To(Equal([]string{"test-events"}))
				Expect(config.Storage).To(Equal(jetstream.FileStorage))
				Expect(config.MaxAge).To(Equal(24 * time.Hour))
			})
		})

		When("using memory storage", func() {
			It("should create config with memory storage", func() {
				publisher.Config.JetStream.Storage = "memory"

				config, err := publisher.createStreamConfig()

				Expect(err).ToNot(HaveOccurred())
				Expect(config.Storage).To(Equal(jetstream.MemoryStorage))
			})
		})

		When("using custom maxAge", func() {
			It("should parse duration correctly", func() {
				publisher.Config.JetStream.MaxAge = "168h" // 7 days in hours

				config, err := publisher.createStreamConfig()

				Expect(err).ToNot(HaveOccurred())
				Expect(config.MaxAge).To(Equal(168 * time.Hour))
			})
		})

		When("using invalid storage type", func() {
			It("should return an error", func() {
				publisher.Config.JetStream.Storage = "invalid"

				_, err := publisher.createStreamConfig()

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid storage type"))
			})
		})

		When("using invalid maxAge duration", func() {
			It("should return an error", func() {
				publisher.Config.JetStream.MaxAge = "invalid-duration"

				_, err := publisher.createStreamConfig()

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("invalid maxAge duration"))
			})
		})
	})

	Context("isStreamNotFoundError", func() {
		When("error is nil", func() {
			It("should return false", func() {
				Expect(isStreamNotFoundError(nil)).To(BeFalse())
			})
		})

		When("error is a JetStream API error with stream not found code", func() {
			It("should return true", func() {
				apiErr := &jetstream.APIError{ErrorCode: jetstream.JSErrCodeStreamNotFound}
				Expect(isStreamNotFoundError(apiErr)).To(BeTrue())
			})
		})

		When("error is standard stream not found error", func() {
			It("should return true", func() {
				err := jetstream.ErrStreamNotFound
				Expect(isStreamNotFoundError(err)).To(BeTrue())
			})
		})

		When("error message matches stream not found patterns", func() {
			It("should return true for various error messages", func() {
				Expect(isStreamNotFoundError(errors.New("nats: stream not found"))).To(BeTrue())
				Expect(isStreamNotFoundError(errors.New("stream not found"))).To(BeTrue())
			})
		})

		When("error is not stream not found", func() {
			It("should return false", func() {
				Expect(isStreamNotFoundError(errors.New("some other error"))).To(BeFalse())
			})
		})
	})
})
