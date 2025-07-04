//go:build unit

package label_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/label"
)

func TestUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Label Unit Suite")
}

var _ = Describe("SanitizeLabels", func() {
	DescribeTable("sanitizes label keys and values",
		func(input map[string]string, expected map[string]string) {
			result := label.SanitizeLabels(input)
			Expect(result).To(Equal(expected))
		},

		Entry("lowercases keys and values",
			map[string]string{"TEST": "TEST"},
			map[string]string{"test": "test"},
		),

		Entry("removes special characters",
			map[string]string{"%^@test*": "%^@test*"},
			map[string]string{"test": "test"},
		),

		Entry("does not change compliant labels",
			map[string]string{"test_test": "test_test"},
			map[string]string{"test_test": "test_test"},
		),

		Entry("truncates long values to 64 characters total",
			func() map[string]string {
				k := "key"
				v := strings.Repeat("v", 80)
				return map[string]string{k: v}
			}(),
			func() map[string]string {
				k := "key"
				v := strings.Repeat("v", 80)
				v = v[:64-len(fmt.Sprintf("%s: ", k))-len("...")] + "..."
				return map[string]string{k: v}
			}(),
		),
	)

	It("returns an empty map when input is empty", func() {
		result := label.SanitizeLabels(map[string]string{})
		Expect(result).To(BeEmpty())
	})
})
