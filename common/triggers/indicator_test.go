//go:build unit

package triggers

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Indicator", func() {
	Describe("AsHeaders", func() {
		It("returns headers only for non-empty fields", func() {
			indicator := Indicator{
				Type:            "onChangePipeline",
				Source:          "source",
				SourceNamespace: "namespace",
			}

			result := indicator.AsHeaders()

			Expect(result).To(HaveKey(Type))
			Expect(result[Type]).To(Equal("trigger-type: onChangePipeline"))

			Expect(result).To(HaveKey(Source))
			Expect(result[Source]).To(Equal("trigger-source: source"))

			Expect(result).To(HaveKey(SourceNamespace))
			Expect(result[SourceNamespace]).To(Equal("trigger-source-namespace: namespace"))
		})

		It("omits empty fields", func() {
			indicator := Indicator{
				Type: "onChangePipeline",
			}

			result := indicator.AsHeaders()

			Expect(result).To(HaveKey(Type))
			Expect(result).NotTo(HaveKey(Source))
			Expect(result).NotTo(HaveKey(SourceNamespace))
		})
	})

	Describe("AsLabels", func() {
		It("returns sanitized labels", func() {
			indicator := Indicator{
				Type:            "onChangePipeline",
				Source:          "source-1",
				SourceNamespace: "namespace/test",
			}

			result := indicator.AsLabels()

			Expect(result).To(HaveKey(TriggerByTypeLabel))
			Expect(result[TriggerByTypeLabel]).To(Equal("onChangePipeline"))

			Expect(result).To(HaveKey(TriggerBySourceLabel))
			Expect(result[TriggerBySourceLabel]).To(Equal("source-1"))

			Expect(result).To(HaveKey(TriggerBySourceNamespaceLabel))
			Expect(result[TriggerBySourceNamespaceLabel]).To(Equal("namespace_test"))
		})

		It("omits empty fields", func() {
			indicator := Indicator{}

			labels := indicator.AsLabels()
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
			// sanitise is unexported, so we'll test it via AsLabels
			indicator := Indicator{
				Type:            "@Type!",
				Source:          "some/source@$",
				SourceNamespace: "some&*Namespace",
			}

			labels := indicator.AsLabels()
			Expect(labels[TriggerByTypeLabel]).To(Equal("Type"))
			Expect(labels[TriggerBySourceLabel]).To(Equal("some_source"))
			Expect(labels[TriggerBySourceNamespaceLabel]).To(Equal("someNamespace"))
		})
	})
})
