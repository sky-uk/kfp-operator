//go:build unit
// +build unit

package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Pipeline", func() {
	var _ = Describe("ComputeHash", func() {

		Specify("Image should change the hash", func() {
			pipeline := Pipeline{}
			hash1 := pipeline.ComputeHash()

			pipeline.Spec.Image = "notempty"
			hash2 := pipeline.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("TfxComponents should change the hash", func() {
			pipeline := Pipeline{}
			hash1 := pipeline.ComputeHash()

			pipeline.Spec.TfxComponents = "notempty"
			hash2 := pipeline.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All Env keys should change the hash", func() {
			pipeline := Pipeline{}
			hash1 := pipeline.ComputeHash()

			pipeline.Spec.Env = map[string]string{
				"a": "",
			}
			hash2 := pipeline.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			pipeline.Spec.Env = map[string]string{
				"b": "notempty",
			}
			hash3 := pipeline.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})

		Specify("All BeamArgs keys should change the hash", func() {
			pipeline := Pipeline{}
			hash1 := pipeline.ComputeHash()

			pipeline.Spec.BeamArgs = map[string]string{
				"a": "",
			}
			hash2 := pipeline.ComputeHash()

			pipeline.Spec.BeamArgs = map[string]string{
				"b": "notempty",
			}
			hash3 := pipeline.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
			Expect(hash2).NotTo(Equal(hash3))
		})
	})

	var _ = Describe("ComputeVersion", func() {

		Specify("Contains the tag if present", func() {
			Expect(Pipeline{Spec: PipelineSpec{
				Image: "image:42",
			}}.ComputeVersion()).To(MatchRegexp("^42-[a-z0-9]{6}$"))

			Expect(Pipeline{Spec: PipelineSpec{
				Image: "docker.io/baz/bar/image:baz",
			}}.ComputeVersion()).To(MatchRegexp("^baz-[a-z0-9]{6}$"))
		})

		Specify("Untagged images should default to latest", func() {
			Expect(Pipeline{Spec: PipelineSpec{
				Image: "image",
			}}.ComputeVersion()).To(MatchRegexp("^latest-[a-z0-9]{6}$"))
		})

		Specify("Malformed image names should have the spec hash only", func() {
			Expect(Pipeline{Spec: PipelineSpec{
				Image: ":",
			}}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})

	var _ = Describe("MarshalJSON", func() {

		Specify("Returns pipeline name if version is missing", func() {
			pid := PipelineIdentifier{Name: "dummy-pipeline"}
			json, err := pid.MarshalJSON()
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(json)).To(Equal("\"dummy-pipeline\""))
		})

		Specify("Returns pipeline name and version if both exist", func() {
			pid := PipelineIdentifier{Name: "dummy-pipeline", Version: "dummy-version"}
			json, err := pid.MarshalJSON()
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(json)).To(Equal("\"dummy-pipeline:dummy-version\""))
		})
	})

	var _ = Describe("UnmarshalJSON", func() {

		Specify("Returns pipeline name if version is missing", func() {
			pid := PipelineIdentifier{Name: "dummy-pipeline"}
			json, err := pid.MarshalJSON()
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(json)).To(Equal("\"dummy-pipeline\""))
		})

		Specify("Returns pipeline name and version if both exist", func() {
			pid := PipelineIdentifier{Name: "dummy-pipeline", Version: "dummy-version"}
			json, err := pid.MarshalJSON()
			Expect(err).To(Not(HaveOccurred()))
			Expect(string(json)).To(Equal("\"dummy-pipeline:dummy-version\""))
		})
	})
})
