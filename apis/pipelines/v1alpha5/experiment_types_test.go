//go:build unit

package v1alpha5

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/pkg/common"
)

var _ = Context("Experiment", func() {
	var _ = Describe("ComputeHash", func() {
		Specify("Description should change the hash", func() {
			rcs := Experiment{}
			hash1 := rcs.ComputeHash()

			rcs.Spec.Description = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("The original object should not change", PropertyBased, func() {
			rcs := RandomExperiment(common.RandomString())
			expected := rcs.DeepCopy()
			rcs.ComputeHash()

			Expect(rcs).To(Equal(expected))
		})
	})

	var _ = Describe("ComputeVersion", func() {
		Specify("Should have the spec hash only", func() {
			Expect(Experiment{}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})
})
