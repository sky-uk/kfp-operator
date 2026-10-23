//go:build unit

package resource

import (
	"encoding/json"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/yaml"
)

var _ = Describe("CompiledPipeline", func() {
	Context("UnmarshalJSON", func() {
		When("given a JSON string", func() {
			It("uses the string contents verbatim, including every YAML document", func() {
				multiDocument := "# PIPELINE DEFINITION\nschemaVersion: 2.1.0\n---\nplatforms:\n  kubernetes: {}\n"
				encoded, err := json.Marshal(multiDocument)
				Expect(err).NotTo(HaveOccurred())

				var compiled CompiledPipeline
				Expect(json.Unmarshal(encoded, &compiled)).To(Succeed())
				Expect(string(compiled)).To(Equal(multiDocument))
			})
		})

		When("given a JSON object", func() {
			It("stores the object as YAML", func() {
				var compiled CompiledPipeline
				Expect(json.Unmarshal([]byte(`{"schemaVersion":"2.1.0","root":{}}`), &compiled)).To(Succeed())

				var got map[string]any
				Expect(yaml.Unmarshal(compiled, &got)).To(Succeed())
				Expect(got).To(Equal(map[string]any{
					"schemaVersion": "2.1.0",
					"root":          map[string]any{},
				}))
			})
		})

		When("given null or no data", func() {
			It("leaves the compiled pipeline nil", func() {
				for _, input := range []string{`null`, ``, `  `} {
					compiled := CompiledPipeline("previous")
					Expect(compiled.UnmarshalJSON([]byte(input))).To(Succeed(), "input %q", input)
					Expect(compiled).To(BeNil(), "input %q", input)
				}
			})
		})

		When("given malformed JSON", func() {
			It("returns an error for an unterminated string", func() {
				var compiled CompiledPipeline
				Expect(compiled.UnmarshalJSON([]byte(`"unterminated`))).To(HaveOccurred())
			})

			It("returns an error for an unterminated object", func() {
				var compiled CompiledPipeline
				Expect(compiled.UnmarshalJSON([]byte(`{"a":`))).To(
					MatchError(ContainSubstring("compiledPipeline must be a YAML string or JSON object")),
				)
			})
		})
	})

	Context("MarshalJSON", func() {
		It("encodes a nil compiled pipeline as null", func() {
			encoded, err := json.Marshal(CompiledPipeline(nil))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(encoded)).To(Equal("null"))
		})

		It("encodes the stored YAML as a JSON string that unmarshals back unchanged", func() {
			original := CompiledPipeline("schemaVersion: 2.1.0\n---\nplatforms:\n  kubernetes: {}\n")

			encoded, err := json.Marshal(original)
			Expect(err).NotTo(HaveOccurred())
			Expect(encoded[0]).To(Equal(byte('"')))

			var roundTripped CompiledPipeline
			Expect(json.Unmarshal(encoded, &roundTripped)).To(Succeed())
			Expect(roundTripped).To(Equal(original))
		})

		It("is omitted from a PipelineDefinitionWrapper when empty", func() {
			encoded, err := json.Marshal(PipelineDefinitionWrapper{})
			Expect(err).NotTo(HaveOccurred())
			Expect(string(encoded)).NotTo(ContainSubstring("compiledPipeline"))
		})
	})
})
