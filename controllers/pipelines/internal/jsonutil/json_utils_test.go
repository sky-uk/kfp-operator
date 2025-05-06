//go:build unit

package jsonutil

import (
	"encoding/json"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	common "github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"testing"
)

func TestJsonUtilUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pipeline Controllers JSON Utils Unit Suite")
}

var _ = Describe("PatchJson", func() {
	When("a json patch is applied", func() {
		It("adds a new field", func() {
			jsonString := "{\"foo\":\"a\",\"bar\":\"b\"}"
			jsonBytes := []byte(jsonString)

			patchOp := common.JsonPatchOperation{
				Op:    "add",
				Path:  "/baz",
				Value: "c",
			}
			patchOpJson, err := json.Marshal([]common.JsonPatchOperation{patchOp})
			Expect(err).To(Not(HaveOccurred()))

			patches := []pipelineshub.Patch{
				{Type: JsonPatch, Payload: string(patchOpJson)},
			}

			result, err := PatchJson(patches, jsonBytes)
			Expect(err).To(Not(HaveOccurred()))

			var data map[string]any
			err = json.Unmarshal([]byte(result), &data)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data["baz"].(string)).To(Equal("c"))
		})

		It("adds to the start of an array field", func() {
			jsonString := "{\"foo\":\"a\",\"bar\":[\"b\"]}"
			jsonBytes := []byte(jsonString)

			patchOp := common.JsonPatchOperation{
				Op:    "add",
				Path:  "/bar/0",
				Value: "newValue",
			}
			patchOpJson, err := json.Marshal([]common.JsonPatchOperation{patchOp})
			Expect(err).To(Not(HaveOccurred()))

			patches := []pipelineshub.Patch{
				{Type: JsonPatch, Payload: string(patchOpJson)},
			}

			result, err := PatchJson(patches, jsonBytes)
			Expect(err).To(Not(HaveOccurred()))

			var data map[string]any
			err = json.Unmarshal([]byte(result), &data)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data["bar"].([]any)[0].(string)).To(Equal("newValue"))
			Expect(data["bar"].([]any)[1].(string)).To(Equal("b"))
		})

		It("creates a new array field and adds one element", func() {
			jsonString := "{\"foo\":\"a\"}"
			jsonBytes := []byte(jsonString)

			patchOp := common.JsonPatchOperation{
				Op:    "add",
				Path:  "/baz/0",
				Value: "hello",
			}
			patchOpJson, err := json.Marshal([]common.JsonPatchOperation{patchOp})
			Expect(err).To(Not(HaveOccurred()))

			patches := []pipelineshub.Patch{
				{Type: JsonPatch, Payload: string(patchOpJson)},
			}

			result, err := PatchJson(patches, jsonBytes)
			Expect(err).To(Not(HaveOccurred()))

			var data map[string]any
			err = json.Unmarshal([]byte(result), &data)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data["baz"].([]any)[0].(string)).To(Equal("hello"))
		})
	})

	When("a merge patch is applied", func() {
		It("merges original json with target json", func() {
			jsonString := "{\"foo\":\"a\",\"bar\":\"b\"}"
			mergeString := "{\"foo\":\"z\",\"baz\":\"c\"}"
			jsonBytes := []byte(jsonString)

			patches := []pipelineshub.Patch{
				{Type: MergePatch, Payload: mergeString},
			}

			result, err := PatchJson(patches, jsonBytes)
			Expect(err).To(Not(HaveOccurred()))

			println(result)
			var data map[string]any
			err = json.Unmarshal([]byte(result), &data)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data["foo"].(string)).To(Equal("z"))
			Expect(data["bar"].(string)).To(Equal("b"))
			Expect(data["baz"].(string)).To(Equal("c"))
		})
	})

	It("fails when using an invalid patch type", func() {
		patches := []pipelineshub.Patch{
			{Type: "invalid", Payload: ""},
		}

		_, err := PatchJson(patches, []byte(""))
		Expect(err).To(HaveOccurred())
	})
})
