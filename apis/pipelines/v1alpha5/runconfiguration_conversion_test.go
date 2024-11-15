//go:build unit

package v1alpha5

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("RunConfiguration Conversion", PropertyBased, func() {
	var _ = Describe("Roundtrip forward", func() {
		Specify("converts to and from the same object using default provider", func() {
			src := RandomRunConfiguration()
			DefaultProvider = "default-provider"
			intermediate := &hub.RunConfiguration{}
			dst := &RunConfiguration{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())
			Expect(getProviderAnnotation(dst)).To(Equal(DefaultProvider))
		})

		Specify("converts to and from the same object", func() {
			src := RandomRunConfiguration()
			src.Status.LatestRuns = LatestRuns{
				Succeeded: RunReference{
					ProviderId: apis.RandomString(),
					Artifacts:  apis.RandomList(common.RandomArtifact),
				},
			}
			setProviderAnnotation(apis.RandomLowercaseString(), &src.ObjectMeta)
			intermediate := &hub.RunConfiguration{}
			dst := &RunConfiguration{}

			Expect(src.ConvertTo(intermediate)).To(Succeed())
			Expect(dst.ConvertFrom(intermediate)).To(Succeed())
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})

	var _ = Describe("Roundtrip backward", func() {
		Specify("converts to and from the same object", func() {
			src := hub.RandomRunConfiguration(apis.RandomLowercaseString())
			hub.WithValueFrom(&src.Spec.Run)
			intermediate := &RunConfiguration{}
			dst := &hub.RunConfiguration{}

			Expect(intermediate.ConvertFrom(src)).To(Succeed())
			Expect(intermediate.ConvertTo(dst)).To(Succeed())
			Expect(dst).To(BeComparableTo(src, cmpopts.EquateEmpty()))
		})
	})
})
