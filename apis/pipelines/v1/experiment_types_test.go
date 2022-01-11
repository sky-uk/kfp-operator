//go:build unit
// +build unit

package v1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Context("Experiment", func() {
	var _ = Describe("ComputeHash", func() {

		Specify("Description should change the hash", func() {
			rcs := ExperimentSpec{}
			hash1 := rcs.ComputeHash()

			rcs.Description = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})
	})

	var _ = Describe("ComputeVersion", func() {

		Specify("Should have the spec hash only", func() {
			Expect(ExperimentSpec{}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})
})
