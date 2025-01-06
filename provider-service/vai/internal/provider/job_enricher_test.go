//go:build unit

package provider

import (
	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ = Describe("JobEnricher", func() {
	var job aiplatformpb.PipelineJob
	var raw map[string]any
	var je = DefaultJobEnricher{}

	BeforeEach(func() {
		job = aiplatformpb.PipelineJob{
			Labels:        map[string]string{"key": "value"},
			RuntimeConfig: &aiplatformpb.PipelineJob_RuntimeConfig{},
		}

		raw = map[string]any{
			"displayName": "test-display-name",
			"pipelineSpec": map[string]any{
				"key": "value",
			},
			"labels": map[string]any{
				"label-key-from-raw": "label-value-from-raw",
			},
			"runtimeConfig": map[string]any{
				"gcsOutputDirectory": "gs://test-bucket",
			},
		}
	})

	Context("Enrich", Ordered, func() {
		When("something", func() {
			It("do something", func() {
				_, err := je.Enrich(&job, raw)
				Expect(err).ToNot(HaveOccurred())
				Expect(job.DisplayName).To(Equal(raw["displayName"]))
				pipelineSpec, err := structpb.NewStruct(
					map[string]any{
						"key": "value",
					},
				)
				Expect(err).ToNot(HaveOccurred())
				Expect(job.PipelineSpec).To(Equal(pipelineSpec))
				Expect(job.Labels).To(Equal(
					map[string]string{
						"key":                "value",
						"label-key-from-raw": "label-value-from-raw",
					},
				))
				Expect(job.RuntimeConfig.GcsOutputDirectory).To(Equal("gs://test-bucket"))
			})
		})
		When("job has no label field", func() {
			It("should set the label field to an empty map", func() {
				job.Labels = nil

				_, err := je.Enrich(&job, raw)
				Expect(err).ToNot(HaveOccurred())
				Expect(job.Labels).To(Equal(map[string]string{
					"label-key-from-raw": "label-value-from-raw",
				}))
			})
		})
		When("displayName is not a string", func() {
			It("should return error", func() {
				raw["displayName"] = 123

				_, err := je.Enrich(&job, raw)
				Expect(err).To(HaveOccurred())
			})
		})
		When("pipelineSpec is not a map", func() {
			It("should return error", func() {
				raw["pipelineSpec"] = 123

				_, err := je.Enrich(&job, raw)
				Expect(err).To(HaveOccurred())
			})
		})
		When("labels is not a map", func() {
			It("should return error", func() {
				raw["labels"] = 123

				_, err := je.Enrich(&job, raw)
				Expect(err).To(HaveOccurred())
			})
		})
		When("runtimeConfig is not a map", func() {
			It("should return error", func() {
				raw["runtimeConfig"] = 123

				_, err := je.Enrich(&job, raw)
				Expect(err).To(HaveOccurred())
			})
		})
		When("runtimeConfig.gcsOutputDirectory is not a string", func() {
			It("should return error", func() {
				raw["runtimeConfig"] = map[string]any{
					"gcsOutputDirectory": 123,
				}

				_, err := je.Enrich(&job, raw)
				Expect(err).To(HaveOccurred())
			})
		})
		When("job has no RuntimeConfig", func() {
			It("should return error", func() {
				job.RuntimeConfig = nil

				_, err := je.Enrich(&job, raw)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
