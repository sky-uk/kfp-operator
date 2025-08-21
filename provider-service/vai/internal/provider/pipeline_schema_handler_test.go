//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ = Describe("SchemaHandler", func() {
	var raw map[string]any
	var schemaHandler = SchemaHandler{}

	BeforeEach(func() {
		raw = map[string]any{
			"displayName": "test-display-name",
			"pipelineSpec": map[string]any{
				"key":           "value",
				"schemaVersion": "2.0.0",
				"sdkVersion":    "tfx-1.15.1",
			},
			"labels": map[string]any{
				"label-key-from-raw": "label-value-from-raw",
			},
			"runtimeConfig": map[string]any{
				"gcsOutputDirectory": "gs://test-bucket",
			},
		}
	})

	Context("Extract", Ordered, func() {
		It("should extract pipeline values from the raw map", func() {
			pipelineValues, err := schemaHandler.extract(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues.name).To(Equal(raw["displayName"]))
			pipelineSpec, err := structpb.NewStruct(
				map[string]any{
					"key":           "value",
					"schemaVersion": "2.0.0",
					"sdkVersion":    "tfx-1.15.1",
				},
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues.pipelineSpec).To(Equal(pipelineSpec))
			Expect(pipelineValues.labels).To(Equal(
				map[string]string{
					"label-key-from-raw": "label-value-from-raw",
					"schema_version":     "2.0.0",
					"sdk_version":        "tfx-1.15.1",
				},
			))
		})

		When("displayName is not a string", func() {
			It("should return error", func() {
				raw["displayName"] = 123

				_, err := schemaHandler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("pipelineSpec is not a map", func() {
			It("should return error", func() {
				raw["pipelineSpec"] = 123

				_, err := schemaHandler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("labels is not a map", func() {
			It("should return error", func() {
				raw["labels"] = 123

				_, err := schemaHandler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("schemaVersion is not a string", func() {
			It("should return error", func() {
				raw["pipelineSpec"].(map[string]any)["schemaVersion"] = 123

				_, err := schemaHandler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("sdkVersion is not a string", func() {
			It("should return error", func() {
				raw["pipelineSpec"].(map[string]any)["sdkVersion"] = 123

				_, err := schemaHandler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
