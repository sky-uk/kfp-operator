//go:build unit

package provider

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ = Describe("SchemaHandler", func() {
	var compiledPipeline resource.CompiledPipeline
	var schemaHandler = SchemaHandler{}

	BeforeEach(func() {
		spec, err := json.Marshal(
			map[string]any{
				"key":           "value",
				"schemaVersion": "2.0.0",
				"sdkVersion":    "tfx-1.15.1",
			},
		)
		Expect(err).ToNot(HaveOccurred())
		compiledPipeline = resource.CompiledPipeline{
			DisplayName: "test-display-name",
			Labels: map[string]string{
				"label-key-from-raw": "label-value-from-raw",
			},
			PipelineSpec: spec,
			// TODO: double check this isn't going to blow things up
			// "runtimeConfig": map[string]any{
			// 	"gcsOutputDirectory": "gs://test-bucket",
			// },
		}
	})

	Context("Extract", Ordered, func() {
		It("should extract pipeline values from the raw map", func() {
			pipelineValues, err := schemaHandler.extract(compiledPipeline)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues.name).To(Equal(compiledPipeline.DisplayName))
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

		// When("pipelineSpec is not a map", func() {
		// 	It("should return error", func() {
		// 		compiledPipeline["pipelineSpec"] = 123
		//
		// 		_, err := schemaHandler.extract(compiledPipeline)
		// 		Expect(err).To(HaveOccurred())
		// 	})
		// })
		//
		// When("schemaVersion is not a string", func() {
		// 	It("should return error", func() {
		// 		compiledPipeline["pipelineSpec"].(map[string]any)["schemaVersion"] = 123
		//
		// 		_, err := schemaHandler.extract(compiledPipeline)
		// 		Expect(err).To(HaveOccurred())
		// 	})
		// })
		//
		// When("sdkVersion is not a string", func() {
		// 	It("should return error", func() {
		// 		compiledPipeline["pipelineSpec"].(map[string]any)["sdkVersion"] = 123
		//
		// 		_, err := schemaHandler.extract(compiledPipeline)
		// 		Expect(err).To(HaveOccurred())
		// 	})
		// })
	})
})
