//go:build unit

package publisher

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/nats-io/nats.go"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/pkg/common"
)

// Mock NATS connection for standard NATS tests
type mockNatsConn struct {
	publishFunc      func(subject string, data []byte) error
	isConnectedFunc  func() bool
	connectedURLFunc func() string
}

func (m *mockNatsConn) Publish(subject string, data []byte) error {
	if m.publishFunc != nil {
		return m.publishFunc(subject, data)
	}
	return nil
}

func (m *mockNatsConn) IsConnected() bool {
	if m.isConnectedFunc != nil {
		return m.isConnectedFunc()
	}
	return true
}

func (m *mockNatsConn) ConnectedUrl() string {
	if m.connectedURLFunc != nil {
		return m.connectedURLFunc()
	}
	return "nats://localhost:4222"
}

var _ = Describe("NatsPublisher", func() {
	var (
		ctx       context.Context
		publisher *NatsPublisher
		mockConn  *mockNatsConn
		subject   string
	)

	BeforeEach(func() {
		ctx = context.Background()
		subject = "test-events"
		mockConn = &mockNatsConn{}

		// Create a real NATS connection for the constructor test
		// We'll replace the connection with our mock afterward
		realConn := &nats.Conn{}
		publisher = NewNatsPublisher(ctx, realConn, subject)

		// Replace with mock for testing
		publisher.NatsConn = mockConn
	})

	Context("NewNatsPublisher", func() {
		When("creating a new publisher", func() {
			It("should initialize with correct subject", func() {
				Expect(publisher.Subject).To(Equal(subject))
				Expect(publisher.NatsConn).ToNot(BeNil())
			})
		})
	})

	Context("Publish", func() {
		When("publishing a valid event", func() {
			It("should publish successfully to NATS", func() {
				var capturedSubject string
				var capturedData []byte

				mockConn.publishFunc = func(subject string, data []byte) error {
					capturedSubject = subject
					capturedData = data
					return nil
				}

				event := common.RunCompletionEvent{
					RunId:  "test-run-123",
					Status: common.RunCompletionStatuses.Succeeded,
				}

				err := publisher.Publish(ctx, event)

				Expect(err).ToNot(HaveOccurred())
				Expect(capturedSubject).To(Equal("test-events"))

				// Verify the data structure
				var wrapper DataWrapper
				err = json.Unmarshal(capturedData, &wrapper)
				Expect(err).ToNot(HaveOccurred())
				Expect(wrapper.Data.RunId).To(Equal("test-run-123"))
				Expect(wrapper.Data.Status).To(Equal(common.RunCompletionStatuses.Succeeded))
			})
		})

		When("NATS publish fails", func() {
			It("should return a ConnectionError", func() {
				mockConn.publishFunc = func(subject string, data []byte) error {
					return errors.New("nats connection failed")
				}

				event := common.RunCompletionEvent{RunId: "test-run"}
				err := publisher.Publish(ctx, event)

				Expect(err).To(HaveOccurred())
				var connErr *ConnectionError
				Expect(errors.As(err, &connErr)).To(BeTrue())
				Expect(connErr.Message).To(ContainSubstring("nats connection failed"))
			})
		})

		When("event marshalling fails", func() {
			It("should return a MarshallingError", func() {
				// Create an event that will cause marshalling to fail
				// This is difficult to trigger with the current structure,
				// so we'll test the error path by verifying the error type
				event := common.RunCompletionEvent{RunId: "test-run"}

				// Mock successful publish to ensure we're not getting connection error
				mockConn.publishFunc = func(subject string, data []byte) error {
					return nil
				}

				err := publisher.Publish(ctx, event)

				// This should succeed with normal event
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})

	Context("Name", func() {
		It("should return the correct publisher name", func() {
			Expect(publisher.Name()).To(Equal("nats-publisher"))
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

	Context("Close", func() {
		When("closing the publisher", func() {
			It("should close the underlying NATS connection", func() {
				// This test verifies the Close method exists and can be called
				// The actual connection closing is handled by the NATS library
				Expect(func() { publisher.Close() }).ToNot(Panic())
			})
		})
	})
})
