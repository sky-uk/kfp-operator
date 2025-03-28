//go:build unit

package v1beta1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

		Specify("Framework should change the hash", func() {
			pipeline := Pipeline{}
			hash1 := pipeline.ComputeHash()

			pipeline.Spec.Framework = PipelineFramework{
				Type: "some-framework",
				Parameters: map[string]*apiextensionsv1.JSON{
					"key": {Raw: []byte(`"value"`)},
				},
			}

			Expect(pipeline.Spec.Framework.Type).To(Equal("some-framework"))

			hash2 := pipeline.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All Env keys should change the hash", func() {
			pipeline := Pipeline{}
			hash1 := pipeline.ComputeHash()

			pipeline.Spec.Env = []apis.NamedValue{
				{Name: "a", Value: ""},
			}
			hash2 := pipeline.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			pipeline.Spec.Env = []apis.NamedValue{
				{Name: "b", Value: "NotEmpty"},
			}
			hash3 := pipeline.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})

		Specify("The original object should not change", PropertyBased, func() {
			rcs := RandomPipeline(common.RandomNamespacedName())
			expected := rcs.DeepCopy()
			rcs.ComputeHash()

			Expect(rcs).To(Equal(expected))
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
