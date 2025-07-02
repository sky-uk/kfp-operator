//go:build unit

package trigger

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
				TriggerType: "type",
			},
			map[string]string{
				TriggerType: "type",
			},
		),

		Entry("trigger-source and trigger-source-namespace present",
			map[string]string{
				TriggerSource:          "github",
				TriggerSourceNamespace: "ci",
			},
			map[string]string{
				TriggerSource:          "github",
				TriggerSourceNamespace: "ci",
			},
		),

		Entry("all headers present",
			map[string]string{
				TriggerType:            "type",
				TriggerSource:          "source",
				TriggerSourceNamespace: "namespace",
			},
			map[string]string{
				TriggerType:            "type",
				TriggerSource:          "source",
				TriggerSourceNamespace: "namespace",
			},
		),

		Entry("irrelevant headers are ignored",
			map[string]string{
				"unrelated": "foo",
				TriggerType: "type",
			},
			map[string]string{
				TriggerType: "type",
			},
		),
	)
})
