//go:build unit

package provider

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LabelService", func() {
	Context("InsertLabelsIntoParameters", func() {
		When("given a list of labels keys, and a compiled pipeline json bytes", func() {
			It("inserts the labels with a default value into the compiled pipeline", func() {
				stubPipelineBytes := []byte(`{}`)
				stubParamDefaultValue := []byte(`{"test": "test"}`)

				labels := []string{"label1", "label2"}

				result, err := DefaultLabelService{
					parameterDefaults: stubParamDefaultValue,
					parameterJsonPath: "testParamPath",
				}.InsertLabelsIntoParameters(stubPipelineBytes, labels)
				Expect(err).ToNot(HaveOccurred())

				Expect(result).To(Equal([]byte(`{"testParamPath":{"label1":{"test": "test"},"label2":{"test": "test"}}}`)))
			})
		})
	})
})
