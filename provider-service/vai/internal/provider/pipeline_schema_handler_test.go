//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/types/known/structpb"
)

var _ = Describe("Schema2_1Handler", func() {
	var (
		raw2_1                       map[string]any
		raw2                         map[string]any
		mockPipelineSchemaHandler2   MockPipelineSchemaHandler
		mockPipelineSchemaHandler2_1 MockPipelineSchemaHandler
		defaultHandler               DefaultPipelineSchemaHandler
		expectedReturn               PipelineValues
	)
	raw2 = map[string]any{
		"pipelineSpec": map[string]any{
			"schemaVersion": "2.0",
		},
	}
	raw2_1 = map[string]any{
		"schemaVersion": "2.1",
	}

	expectedReturn = PipelineValues{}

	BeforeEach(func() {
		mockPipelineSchemaHandler2 = MockPipelineSchemaHandler{}
		mockPipelineSchemaHandler2_1 = MockPipelineSchemaHandler{}
		defaultHandler = DefaultPipelineSchemaHandler{
			schema2Handler:   &mockPipelineSchemaHandler2,
			schema2_1Handler: &mockPipelineSchemaHandler2_1,
		}
	})

	Context("Extract", Ordered, func() {
		It("should use 2.0 handler", func() {
			mockPipelineSchemaHandler2.On("extract", raw2).Return(&expectedReturn, nil)
			pipelineValues, err := defaultHandler.extract(raw2)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues).To(Equal(&expectedReturn))
			Expect(mockPipelineSchemaHandler2_1.Calls).To(BeEmpty())
		})

		It("should use 2.1 handler", func() {
			mockPipelineSchemaHandler2_1.On("extract", raw2_1).Return(&expectedReturn, nil)
			pipelineValues, err := defaultHandler.extract(raw2_1)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues).To(Equal(&expectedReturn))
			Expect(mockPipelineSchemaHandler2.Calls).To(BeEmpty())
		})

		It("should return error when schemaVersion is not found", func() {
			_, err := defaultHandler.extract(map[string]any{})
			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("Schema2_1Handler", func() {
	var raw map[string]any
	var schema21Handler = Schema2_1Handler{}

	BeforeEach(func() {
		raw = map[string]any{
			"pipelineInfo": map[string]any{
				"name": "test-display-name",
			},
			"sdkVersion":    "KFP-1.2.3",
			"schemaVersion": "Version-1.2.3",
			"components": map[string]any{
				"component1": "some-component",
			},
			"deploymentSpec": map[string]any{
				"executors": "some-executor",
			},
			"root": map[string]any{
				"dag": "some-dag",
			},
		}
	})

	Context("Extract", Ordered, func() {
		It("should return the pipeline values from the raw map", func() {
			pipelineValues, err := schema21Handler.extract(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues.name).To(Equal(raw["pipelineInfo"].(map[string]any)["name"].(string)))
			Expect(pipelineValues.labels).To(Equal(
				map[string]string{
					"sdk_version":    "kfp-1_2_3",
					"schema_version": "version-1_2_3",
				},
			))
			deploySpec := pipelineValues.pipelineSpec.AsMap()["deploymentSpec"].(map[string]any)
			components := pipelineValues.pipelineSpec.AsMap()["components"].(map[string]any)
			root := pipelineValues.pipelineSpec.AsMap()["root"].(map[string]any)
			Expect(components["component1"]).To(Equal("some-component"))
			Expect(deploySpec["executors"]).To(Equal("some-executor"))
			Expect(root["dag"]).To(Equal("some-dag"))
		})

		When("pipelineInfo is not a set", func() {
			It("should return error", func() {
				raw["pipelineInfo"] = nil

				_, err := schema21Handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("displayName is not a string", func() {
			It("should return error", func() {
				raw["pipelineInfo"].(map[string]any)["name"] = 123

				_, err := schema21Handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("schemaVersion is not a string", func() {
			It("should return error", func() {
				raw["schemaVersion"] = 123

				_, err := schema21Handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})

			It("should return error", func() {
				raw["sdkVersion"] = 123

				_, err := schema21Handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})

var _ = Describe("Schema2Handler", func() {
	var raw map[string]any
	var schema2Handler = Schema2Handler{}

	BeforeEach(func() {
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

	Context("Extract", Ordered, func() {
		It("should extract pipeline values from the raw map", func() {
			pipelineValues, err := schema2Handler.extract(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues.name).To(Equal(raw["displayName"]))
			pipelineSpec, err := structpb.NewStruct(
				map[string]any{
					"key": "value",
				},
			)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues.pipelineSpec).To(Equal(pipelineSpec))
			Expect(pipelineValues.labels).To(Equal(
				map[string]string{
					"label-key-from-raw": "label-value-from-raw",
				},
			))
		})
		When("displayName is not a string", func() {
			It("should return error", func() {
				raw["displayName"] = 123

				_, err := schema2Handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})
		When("pipelineSpec is not a map", func() {
			It("should return error", func() {
				raw["pipelineSpec"] = 123

				_, err := schema2Handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})
		When("labels is not a map", func() {
			It("should return error", func() {
				raw["labels"] = 123

				_, err := schema2Handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
