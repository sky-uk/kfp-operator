//go:build unit

package provider

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"errors"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
)

type MockPipelineSchemaHandler struct{ mock.Mock }

func (m *MockPipelineSchemaHandler) extract(raw map[string]any) (*PipelineValues, error) {
	args := m.Called(raw)
	var pipelineValues *PipelineValues
	if arg0 := args.Get(0); arg0 != nil {
		pipelineValues = arg0.(*PipelineValues)
	}
	return pipelineValues, args.Error(1)
}

var _ = Describe("DefaultJobEnricher", func() {

	var (
		mockPipelineSchemaHandler MockPipelineSchemaHandler
		dje                       DefaultJobEnricher
		expectedReturn            PipelineValues
	)

	BeforeEach(func() {
		mockPipelineSchemaHandler = MockPipelineSchemaHandler{}
		dje = DefaultJobEnricher{pipelineSchemaHandler: &mockPipelineSchemaHandler}
		expectedReturn = PipelineValues{name: "enriched", labels: map[string]string{"key": "value"}, pipelineSpec: &structpb.Struct{}}
	})

	Context("Enrich", Ordered, func() {
		input := map[string]any{"schemaVersion": "2.0"}
		It("enrich job with pipeline values returned by pipelineSchemaHandler", func() {
			mockPipelineSchemaHandler.On("extract", input).Return(&expectedReturn, nil)

			job := aiplatformpb.PipelineJob{}
			_, err := dje.Enrich(&job, input)
			Expect(err).ToNot(HaveOccurred())

			Expect(job.Name).To(Equal(expectedReturn.name))
			Expect(job.Labels).To(Equal(expectedReturn.labels))
			Expect(job.PipelineSpec).To(Equal(expectedReturn.pipelineSpec))
		})

		It("enrich job returns error on pipelineSchemaHandler error", func() {
			mockPipelineSchemaHandler.On("extract", input).Return(nil, errors.New("an error"))

			job := aiplatformpb.PipelineJob{}
			_, err := dje.Enrich(&job, input)
			Expect(err).To(HaveOccurred())
		})
	})
})
