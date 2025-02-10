//go:build unit

package pipelines

import (
	"encoding/json"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Provider Deployment Manager", func() {

	var _ = Describe("jsonToString", func() {
		Specify("Should return a plain string given a JSON string (no extra quotes or escape chars!)", func() {
			rawJson, err := json.Marshal("test")
			Expect(err).ToNot(HaveOccurred())

			jsonStr := apiextensionsv1.JSON{
				Raw: rawJson,
			}

			result := jsonToString(&jsonStr)

			Expect(result).To(Equal("test"))
		})

		Specify("Should return a raw JSON string given a JSON object", func() {
			rawJson, err := json.Marshal(`{"key1": "value1", "key2": 42}`)
			Expect(err).ToNot(HaveOccurred())

			jsonStr := apiextensionsv1.JSON{
				Raw: rawJson,
			}

			result := jsonToString(&jsonStr)

			Expect(result).To(Equal("{\"key1\": \"value1\", \"key2\": 42}"))
		})
	})

})
