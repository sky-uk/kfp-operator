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
		It("adding a new field", func() {
			jsonString := "{\"foo\":\"a\",\"bar\":\"b\"}"
			jsonBytes := []byte(jsonString)

			patchOp := common.PatchOperation{
				Op:    "add",
				Path:  "/baz",
				Value: "c",
			}
			patchOpJson, err := json.Marshal([]common.PatchOperation{patchOp})
			Expect(err).To(Not(HaveOccurred()))

			patches := []pipelineshub.Patch{
				{Type: JsonPatch, Patch: string(patchOpJson)},
			}

			result, err := PatchJson(patches, jsonBytes)
			Expect(err).To(Not(HaveOccurred()))

			var data map[string]interface{}
			err = json.Unmarshal([]byte(result), &data)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data["baz"].(string)).To(Equal("c"))
		})

		It("adding to the start of an array field", func() {
			jsonString := "{\"foo\":\"a\",\"bar\":[\"b\"]}"
			jsonBytes := []byte(jsonString)

			patchOp := common.PatchOperation{
				Op:    "add",
				Path:  "/bar/0",
				Value: "newValue",
			}
			patchOpJson, err := json.Marshal([]common.PatchOperation{patchOp})
			Expect(err).To(Not(HaveOccurred()))

			patches := []pipelineshub.Patch{
				{Type: JsonPatch, Patch: string(patchOpJson)},
			}

			result, err := PatchJson(patches, jsonBytes)
			Expect(err).To(Not(HaveOccurred()))

			var data map[string]interface{}
			err = json.Unmarshal([]byte(result), &data)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data["bar"].([]interface{})[0].(string)).To(Equal("newValue"))
			Expect(data["bar"].([]interface{})[1].(string)).To(Equal("b"))
		})

		It("create a new array field and add one element", func() {
			jsonString := "{\"foo\":\"a\"}"
			jsonBytes := []byte(jsonString)

			patchOp := common.PatchOperation{
				Op:    "add",
				Path:  "/baz/0",
				Value: "hello",
			}
			patchOpJson, err := json.Marshal([]common.PatchOperation{patchOp})
			Expect(err).To(Not(HaveOccurred()))

			patches := []pipelineshub.Patch{
				{Type: JsonPatch, Patch: string(patchOpJson)},
			}

			result, err := PatchJson(patches, jsonBytes)
			Expect(err).To(Not(HaveOccurred()))

			var data map[string]interface{}
			err = json.Unmarshal([]byte(result), &data)
			Expect(err).To(Not(HaveOccurred()))
			Expect(data["baz"].([]interface{})[0].(string)).To(Equal("hello"))
		})

		It("fails when using an invalid patch type", func() {
			patches := []pipelineshub.Patch{
				{Type: "invalid", Patch: ""},
			}

			_, err := PatchJson(patches, []byte(""))
			Expect(err).To(HaveOccurred())
		})
	})
})

// add tests for MergePatch
