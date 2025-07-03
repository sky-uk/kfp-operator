//go:build unit

package triggers

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTriggerUnitSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Triggers Unit Test")
}

var _ = Describe("FromHeaders", func() {
	DescribeTable("extracts correct trigger headers",
		func(input map[string]string, expected map[string]string) {
			result := FromHeaders(input)
			Expect(result).To(Equal(expected))
		},

		Entry("empty headers",
			map[string]string{},
			map[string]string{},
		),

		Entry("only trigger-type present",
			map[string]string{
				Type: "type",
			},
			map[string]string{
				Type: "type",
			},
		),

		Entry("trigger-source and trigger-source-namespace present",
			map[string]string{
				Source:          "github",
				SourceNamespace: "ci",
			},
			map[string]string{
				Source:          "github",
				SourceNamespace: "ci",
			},
		),

		Entry("all headers present",
			map[string]string{
				Type:            "type",
				Source:          "source",
				SourceNamespace: "namespace",
			},
			map[string]string{
				Type:            "type",
				Source:          "source",
				SourceNamespace: "namespace",
			},
		),

		Entry("irrelevant headers are ignored",
			map[string]string{
				"unrelated": "foo",
				Type:        "type",
			},
			map[string]string{
				Type: "type",
			},
		),
	)
})
