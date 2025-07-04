//go:build unit

package triggers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Indicator", func() {
	Describe("AsK8sLabels", func() {
		It("returns sanitized labels", func() {
			indicator := Indicator{
				Type:            "onChangePipeline",
				Source:          "source-1",
				SourceNamespace: "namespace/test",
			}

			result := indicator.AsK8sLabels()

			Expect(result).To(HaveKey(TriggerByTypeLabel))
			Expect(result[TriggerByTypeLabel]).To(Equal("onChangePipeline"))

			Expect(result).To(HaveKey(TriggerBySourceLabel))
			Expect(result[TriggerBySourceLabel]).To(Equal("source-1"))

			Expect(result).To(HaveKey(TriggerBySourceNamespaceLabel))
			Expect(result[TriggerBySourceNamespaceLabel]).To(Equal("namespace_test"))
		})

		It("omits empty fields", func() {
			indicator := Indicator{}

			labels := indicator.AsK8sLabels()
			Expect(labels).To(BeEmpty())
		})
	})

	Describe("FromLabels", func() {
		It("builds indicator from label map", func() {
			labels := map[string]string{
				TriggerByTypeLabel:            "onChangeRunSpec",
				TriggerBySourceLabel:          "source",
				TriggerBySourceNamespaceLabel: "namespace",
			}

			result := FromLabels(labels)

			Expect(result.Type).To(Equal("onChangeRunSpec"))
			Expect(result.Source).To(Equal("source"))
			Expect(result.SourceNamespace).To(Equal("namespace"))
		})

		It("returns empty fields when labels are missing", func() {
			result := FromLabels(map[string]string{})

			Expect(result.Type).To(BeEmpty())
			Expect(result.Source).To(BeEmpty())
			Expect(result.SourceNamespace).To(BeEmpty())
		})
	})

	Describe("sanitise", func() {
		It("removes invalid characters and replaces slashes", func() {
			// sanitise is unexported, so we'll test it via AsK8sLabels
			indicator := Indicator{
				Type:            "@Type!",
				Source:          "some/source@$",
				SourceNamespace: "some&*Namespace",
			}

			labels := indicator.AsK8sLabels()
			Expect(labels[TriggerByTypeLabel]).To(Equal("Type"))
			Expect(labels[TriggerBySourceLabel]).To(Equal("some_source"))
			Expect(labels[TriggerBySourceNamespaceLabel]).To(Equal("someNamespace"))
		})
	})
})
