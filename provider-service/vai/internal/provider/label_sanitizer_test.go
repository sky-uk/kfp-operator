//go:build unit

package provider

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("DefaultLabelSanitizer", func() {
	dls := DefaultLabelSanitizer{}

	Context("Sanitize", func() {
		DescribeTable(
			"sanitizes label keys and values",
			func(input map[string]string, expected map[string]string) {
				result := dls.Sanitize(input)
				Expect(result).To(Equal(expected))
			},

			Entry(
				"lowercases keys and values",
				map[string]string{"TEST": "TEST"},
				map[string]string{"test": "test"},
			),

			Entry(
				"removes special characters",
				map[string]string{"%^@test*": "%^@test*"},
				map[string]string{"test": "test"},
			),

			Entry(
				"does not change compliant labels",
				map[string]string{"test_test": "test_test"},
				map[string]string{"test_test": "test_test"},
			),

			Entry(
				"if key is pipeline-version, schema_version or sdk_version then it replaces invalid characters with underscore",
				map[string]string{"pipeline-version": "0.0.1", "schema_version": "2.1.0", "sdk_version": "kfp-2.12.2"},
				map[string]string{"pipeline-version": "0_0_1", "schema_version": "2_1_0", "sdk_version": "kfp-2_12_2"},
			),
		)

		It("trims keys and values to 63 characters", func() {
			maxLength := 63

			key := strings.Repeat("k", 100)
			value := strings.Repeat("v", 100)

			result := dls.Sanitize(map[string]string{key: value})
			Expect(result).To(Equal(map[string]string{
				key[:maxLength]: value[:maxLength],
			}))
			Expect(len(key[:maxLength])).To(Equal(63))
		})

		It("returns an empty map when input is empty", func() {
			result := dls.Sanitize(map[string]string{})
			Expect(result).To(BeEmpty())
		})
	})
})
