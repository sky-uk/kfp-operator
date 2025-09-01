//go:build unit

package provider

import (
	"encoding/json"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/mocks"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ = Describe("DefaultJobEnricher", func() {
	Context("Enrich", Ordered, func() {
		var (
			labelSanitizer     mocks.MockLabelSanitizer
			defaultJobEnricher DefaultJobEnricher
			input              resource.CompiledPipeline
		)

		spec := map[string]string{
			"schemaVersion": "2.1.0",
			"sdkVersion":    "kfp-2.12.2",
		}

		specBytes, err := json.Marshal(spec)
		Expect(err).ToNot(HaveOccurred())

		BeforeEach(func() {
			labelSanitizer = mocks.MockLabelSanitizer{}
			defaultJobEnricher = DefaultJobEnricher{
				labelSanitizer: &labelSanitizer,
			}

			input = resource.CompiledPipeline{
				DisplayName: "display-name",
				Labels: map[string]string{
					"key":            "value",
					"tfx_py_version": "3-10",
					"tfx_runner":     "kubeflow_v2",
					"tfx_version":    "1-15-1",
				},
				PipelineSpec: specBytes,
			}
		})

		It("enriches job with displayName, sanitized labels and PipelineSpec", func() {
			job := aiplatformpb.PipelineJob{}
			combinedLabels := map[string]string{
				"key":            "value",
				"tfx_py_version": "3-10",
				"tfx_runner":     "kubeflow_v2",
				"tfx_version":    "1-15-1",
				"schema_version": "2.1.0",
				"sdk_version":    "kfp-2.12.2",
			}
			sanitizedLabels := map[string]string{
				"key":            "value",
				"tfx_py_version": "3-10",
				"tfx_runner":     "kubeflow_v2",
				"tfx_version":    "1-15-1",
				"schema_version": "2-1-0",
				"sdk_version":    "kfp-2-12-2",
			}

			labelSanitizer.On("Sanitize", combinedLabels).Return(sanitizedLabels)

			castedSpec := make(map[string]any, len(spec))
			for k, v := range spec {
				castedSpec[k] = v
			}

			pipelineSpecPb, err := structpb.NewStruct(castedSpec)
			Expect(err).ToNot(HaveOccurred())

			_, err = defaultJobEnricher.Enrich(&job, input)
			Expect(err).ToNot(HaveOccurred())

			Expect(job.Name).To(Equal(input.DisplayName))
			Expect(job.Labels).To(Equal(sanitizedLabels))
			Expect(job.PipelineSpec).To(Equal(pipelineSpecPb))
		})

		When("schemaVersion key is missing from PipelineSpec", func() {
			It("returns error", func() {
				spec, err := json.Marshal(
					map[string]string{"sdkVersion": "kfp-2.12.2"},
				)
				Expect(err).ToNot(HaveOccurred())

				job := aiplatformpb.PipelineJob{}
				input.PipelineSpec = spec
				_, err = defaultJobEnricher.Enrich(&job, input)
				Expect(err).To(HaveOccurred())
			})
		})

		When("sdkVersion key is missing from PipelineSpec", func() {
			It("returns error", func() {
				spec, err := json.Marshal(
					map[string]string{"schemaVersion": "2.1.0"},
				)
				Expect(err).ToNot(HaveOccurred())

				job := aiplatformpb.PipelineJob{}
				input.PipelineSpec = spec
				_, err = defaultJobEnricher.Enrich(&job, input)
				Expect(err).To(HaveOccurred())
			})
		})

		When("PipelineSpec cannot be unmarshalled", func() {
			It("returns error", func() {
				input.PipelineSpec = []byte{'1'}
				job := aiplatformpb.PipelineJob{}
				_, err := defaultJobEnricher.Enrich(&job, input)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
