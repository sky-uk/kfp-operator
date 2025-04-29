//go:build unit

package publisher

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/stretchr/testify/mock"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "PublisherHandler Unit Test Suite")
}

type MockNatsConn struct {
	mock.Mock
}

func (m *MockNatsConn) Publish(subject string, data []byte) error {
	args := m.Called(subject, data)
	return args.Error(0)
}

func (m *MockNatsConn) IsConnected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNatsConn) IsClosed() bool {
	args := m.Called()
	return args.Bool(0)
}

var _ = Describe("PublisherHandler", func() {
	var (
		mockNatsConn *MockNatsConn
		publisher    *NatsPublisher
		subject      string
		event        common.RunCompletionEvent
	)

	BeforeEach(func() {
		mockNatsConn = &MockNatsConn{}
		subject = "test.subject"
		publisher = &NatsPublisher{
			NatsConn: mockNatsConn,
			Subject:  subject,
		}
		event = common.RunCompletionEvent{}
	})

	Context("Publish", func() {
		When("run completion event marshalling and NatsConn.Publish are successful", func() {
			It("should not error", func() {
				eventData, _ := json.Marshal(DataWrapper{Data: event})
				mockNatsConn.On("Publish", subject, mock.MatchedBy(func(data []byte) bool {
					return bytes.Equal(data, eventData)
				})).Return(nil)

				err := publisher.Publish(event)

				Expect(err).To(Not(HaveOccurred()))
				mockNatsConn.AssertExpectations(GinkgoT())
			})
		})

		When("NatsConn.Publish errors", func() {
			It("should return an error", func() {
				dataWrapper := DataWrapper{Data: event}
				eventData, _ := json.Marshal(dataWrapper)
				mockNatsConn.On("Publish", subject, mock.MatchedBy(func(data []byte) bool {
					return bytes.Equal(data, eventData)
				})).Return(errors.New("Connection failed"))

				err := publisher.Publish(event)

				Expect(err).To(BeAssignableToTypeOf(&ConnectionError{}))
				Expect(err.Error()).To(ContainSubstring("Connection failed"))
			})
		})
	})
})
