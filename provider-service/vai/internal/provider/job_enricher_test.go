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
	)

	expectedReturn := PipelineValues{
		name: "enriched",
		labels: map[string]string{
			"key":            "value",
			"schema_version": "2.1.0",
			"sdk_version":    "kfp-2.12.2",
			"Other&%":        "vAl!ue2",
		},
		pipelineSpec: &structpb.Struct{},
	}

	BeforeEach(func() {
		mockPipelineSchemaHandler = MockPipelineSchemaHandler{}
		dje = DefaultJobEnricher{pipelineSchemaHandler: &mockPipelineSchemaHandler}
	})

	Context("Enrich", Ordered, func() {
		input := map[string]any{"schemaVersion": "2.0"}
		It("enriches job with labels returned by pipelineSchemaHandler which are sanitized", func() {
			mockPipelineSchemaHandler.On("extract", input).Return(&expectedReturn, nil)

			sanitizedLabels := map[string]string{
				"key":            "value",
				"schema_version": "2_1_0",
				"sdk_version":    "kfp-2_12_2",
				"other":          "value2",
			}

			job := aiplatformpb.PipelineJob{}
			_, err := dje.Enrich(&job, input)
			Expect(err).ToNot(HaveOccurred())

			Expect(job.Name).To(Equal(expectedReturn.name))
			Expect(job.Labels).To(Equal(sanitizedLabels))
			Expect(job.PipelineSpec).To(Equal(expectedReturn.pipelineSpec))
		})

		It("combines and sanitizes job labels and labels returned by pipelineSchemaHandler", func() {
			mockPipelineSchemaHandler.On("extract", input).Return(&expectedReturn, nil)

			expectedCombinedLabels := map[string]string{
				"key":            "value",
				"schema_version": "2_1_0",
				"sdk_version":    "kfp-2_12_2",
				"other":          "value2",
				"key2":           "value2",
			}

			job := aiplatformpb.PipelineJob{Labels: map[string]string{"key2": "value2"}}
			_, err := dje.Enrich(&job, input)
			Expect(err).ToNot(HaveOccurred())
			Expect(job.Name).To(Equal(expectedReturn.name))
			Expect(job.Labels).To(Equal(expectedCombinedLabels))
			Expect(job.PipelineSpec).To(Equal(expectedReturn.pipelineSpec))
		})

		It("returns error on pipelineSchemaHandler error", func() {
			mockPipelineSchemaHandler.On("extract", input).Return(nil, errors.New("an error"))

			job := aiplatformpb.PipelineJob{}
			_, err := dje.Enrich(&job, input)
			Expect(err).To(HaveOccurred())
		})
	})
})
