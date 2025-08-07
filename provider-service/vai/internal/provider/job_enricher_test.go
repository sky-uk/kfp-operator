//go:build unit

package provider

import (
	"errors"
	"strings"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
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
	Context("Enrich", Ordered, func() {
		var (
			mockPipelineSchemaHandler MockPipelineSchemaHandler
			dje                       DefaultJobEnricher
		)

		expectedReturn := PipelineValues{
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
			mockPipelineSchemaHandler = MockPipelineSchemaHandler{}
			dje = DefaultJobEnricher{pipelineSchemaHandler: &mockPipelineSchemaHandler}
		})
		input := map[string]any{"schemaVersion": "2.0"}
		It("enriches job with labels returned by pipelineSchemaHandler which are sanitized", func() {
			mockPipelineSchemaHandler.On("extract", input).Return(&expectedReturn, nil)

			sanitizedLabels := map[string]string{
				"key":              "value",
				"pipeline-version": "0-0-1",
				"schema_version":   "2_1_0",
				"sdk_version":      "kfp-2_12_2",
				"other":            "somevalue",
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
				"key":              "value",
				"pipeline-version": "0-0-1",
				"schema_version":   "2_1_0",
				"sdk_version":      "kfp-2_12_2",
				"other":            "somevalue",
				"key2":             "value2",
			}

			job := aiplatformpb.PipelineJob{Labels: map[string]string{"Key2": "Value2"}}
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

	Context("sanitizeLabels", func() {
		DescribeTable(
			"sanitizes label keys and values",
			func(input map[string]string, expected map[string]string) {
				result := sanitizeLabels(input)
				Expect(result).To(Equal(expected))
			},

			Entry(
				"lowercases keys and values",
				map[string]string{"TEST": "TEST"},
				map[string]string{"test": "test"},
			),

			Entry(
				"removes special characters",
				map[string]string{"%^@test*": "%^@test*"},
				map[string]string{"test": "test"},
			),

			Entry(
				"does not change compliant labels",
				map[string]string{"test_test": "test_test"},
				map[string]string{"test_test": "test_test"},
			),

			Entry(
				"if key is pipeline-version then it replaces invalid characters with hyphen",
				map[string]string{"pipeline-version": "0.0.1"},
				map[string]string{"pipeline-version": "0-0-1"},
			),

			Entry(
				"if key is schema_version or sdk_version then it replaces invalid characters with underscore",
				map[string]string{"schema_version": "2.1.0", "sdk_version": "kfp-2.12.2"},
				map[string]string{"schema_version": "2_1_0", "sdk_version": "kfp-2_12_2"},
			),
		)

		It("trims keys and values to 63 characters", func() {
			maxLength := 63

			key := strings.Repeat("k", 100)
			value := strings.Repeat("v", 100)

			result := sanitizeLabels(map[string]string{key: value})
			Expect(result).To(Equal(map[string]string{
				key[:maxLength]: value[:maxLength],
			}))
			Expect(len(key[:maxLength])).To(Equal(63))
		})

		It("returns an empty map when input is empty", func() {
			result := sanitizeLabels(map[string]string{})
			Expect(result).To(BeEmpty())
		})
	})
})
