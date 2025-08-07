//go:build unit

package provider

import (
	"errors"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/mocks"
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
	Context("Enrich", Ordered, func() {
		var (
			pipelineSchemeHandler MockPipelineSchemaHandler
			labelSanitizer        mocks.MockLabelSanitizer
			dje                   DefaultJobEnricher
		)

		pipelineValues := PipelineValues{
			name: "enriched",
			labels: map[string]string{
				"key":              "value",
				"pipeline-version": "0.0.1",
				"schema_version":   "2.1.0",
				"sdk_version":      "kfp-2.12.2",
				"Other&%":          "someVAl!ue",
			},
			pipelineSpec: &structpb.Struct{},
		}

		BeforeEach(func() {
			pipelineSchemeHandler = MockPipelineSchemaHandler{}
			labelSanitizer = mocks.MockLabelSanitizer{}
			dje = DefaultJobEnricher{
				pipelineSchemaHandler: &pipelineSchemeHandler,
				labelSanitizer:        &labelSanitizer,
			}
		})

		input := map[string]any{"somekey": "somevalue"}

		It("enriches job with labels returned by pipelineSchemaHandler", func() {
			pipelineSchemeHandler.On("extract", input).Return(&pipelineValues, nil)
			labelSanitizer.On("Sanitize", pipelineValues.labels).Return(pipelineValues.labels)

			job := aiplatformpb.PipelineJob{}
			_, err := dje.Enrich(&job, input)
			Expect(err).ToNot(HaveOccurred())

			Expect(job.Name).To(Equal(pipelineValues.name))
			Expect(job.Labels).To(Equal(pipelineValues.labels))
			Expect(job.PipelineSpec).To(Equal(pipelineValues.pipelineSpec))
		})

		It("combines job labels and labels returned by pipelineSchemaHandler", func() {
			pipelineSchemeHandler.On("extract", input).Return(&pipelineValues, nil)
			combinedLabels := map[string]string{
				"key":              "value",
				"pipeline-version": "0.0.1",
				"schema_version":   "2.1.0",
				"sdk_version":      "kfp-2.12.2",
				"Other&%":          "someVAl!ue",
				"Key2":             "Value2",
			}
			labelSanitizer.On("Sanitize", combinedLabels).Return(combinedLabels)

			job := aiplatformpb.PipelineJob{Labels: map[string]string{"Key2": "Value2"}}
			_, err := dje.Enrich(&job, input)
			Expect(err).ToNot(HaveOccurred())
			Expect(job.Name).To(Equal(pipelineValues.name))
			Expect(job.Labels).To(Equal(combinedLabels))
			Expect(job.PipelineSpec).To(Equal(pipelineValues.pipelineSpec))
		})

		It("returns error on pipelineSchemaHandler error", func() {
			pipelineSchemeHandler.On("extract", input).Return(nil, errors.New("an error"))

			job := aiplatformpb.PipelineJob{}
			_, err := dje.Enrich(&job, input)
			Expect(err).To(HaveOccurred())
		})
	})
})
