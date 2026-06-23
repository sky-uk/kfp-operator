//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DefaultPipelineSchemaHandler", func() {
	var handler = DefaultPipelineSchemaHandler{}

	Context("wrapped IR (Vertex PipelineJob envelope)", func() {
		var raw map[string]any

		BeforeEach(func() {
			raw = map[string]any{
				"displayName": "test-display-name",
				"pipelineSpec": map[string]any{
					"key":           "value",
					"schemaVersion": "2.0.0",
					"sdkVersion":    "tfx-1.15.1",
					"pipelineInfo": map[string]any{
						"name": "inner-pipeline-name",
					},
				},
				"labels": map[string]any{
					"label-key-from-raw": "label-value-from-raw",
				},
				"runtimeConfig": map[string]any{
					"gcsOutputDirectory": "gs://test-bucket",
				},
			}
		})

		It("uses the wrapper displayName and the inner spec for a 2.0 spec", func() {
			pipelineValues, err := handler.extract(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues.name).To(Equal("test-display-name"))
			Expect(pipelineValues.labels).To(Equal(map[string]string{
				"label-key-from-raw": "label-value-from-raw",
				"schema_version":     "2.0.0",
				"sdk_version":        "tfx-1.15.1",
			}))
			specMap := pipelineValues.pipelineSpec.AsMap()
			Expect(specMap["key"]).To(Equal("value"))
			Expect(specMap).ToNot(HaveKey("pipelineSpec"))
		})

		It("uses the wrapper displayName for a wrapped 2.1 spec (regression)", func() {
			raw["pipelineSpec"].(map[string]any)["schemaVersion"] = "2.1.0"
			raw["pipelineSpec"].(map[string]any)["sdkVersion"] = "kfp-2.0.1"
			pipelineValues, err := handler.extract(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues.name).To(Equal("test-display-name"))
			Expect(pipelineValues.labels["schema_version"]).To(Equal("2.1.0"))
			Expect(pipelineValues.labels["sdk_version"]).To(Equal("kfp-2.0.1"))
		})

		When("schemaVersion in the inner spec is not a string", func() {
			It("should return error", func() {
				raw["pipelineSpec"].(map[string]any)["schemaVersion"] = 123

				_, err := handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("sdkVersion in the inner spec is not a string", func() {
			It("should return error", func() {
				raw["pipelineSpec"].(map[string]any)["sdkVersion"] = 123

				_, err := handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Context("bare IR (kfp-sdk output)", func() {
		var raw map[string]any

		BeforeEach(func() {
			raw = map[string]any{
				"pipelineInfo": map[string]any{
					"name": "test-pipeline-name",
				},
				"sdkVersion":    "kfp-2.0.1",
				"schemaVersion": "2.1.0",
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

		It("uses pipelineInfo.name and the bare spec", func() {
			pipelineValues, err := handler.extract(raw)
			Expect(err).ToNot(HaveOccurred())
			Expect(pipelineValues.name).To(Equal("test-pipeline-name"))
			Expect(pipelineValues.labels).To(Equal(map[string]string{
				"sdk_version":    "kfp-2.0.1",
				"schema_version": "2.1.0",
			}))
			specMap := pipelineValues.pipelineSpec.AsMap()
			Expect(specMap["components"].(map[string]any)["component1"]).To(Equal("some-component"))
			Expect(specMap["deploymentSpec"].(map[string]any)["executors"]).To(Equal("some-executor"))
			Expect(specMap["root"].(map[string]any)["dag"]).To(Equal("some-dag"))
		})

		When("pipelineInfo is not set", func() {
			It("should return error", func() {
				raw["pipelineInfo"] = nil

				_, err := handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("pipelineInfo.name is not a string", func() {
			It("should return error", func() {
				raw["pipelineInfo"].(map[string]any)["name"] = 123

				_, err := handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("schemaVersion is not a string", func() {
			It("should return error", func() {
				raw["schemaVersion"] = 123

				_, err := handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})

		When("sdkVersion is not a string", func() {
			It("should return error", func() {
				raw["sdkVersion"] = 123

				_, err := handler.extract(raw)
				Expect(err).To(HaveOccurred())
			})
		})
	})
})
