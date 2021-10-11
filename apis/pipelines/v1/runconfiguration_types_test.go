//go:build unit
// +build unit

package v1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Context("RunConfiguration", func() {
	var _ = Describe("ComputeHash", func() {

		Specify("PipelineName should change the hash", func() {
			rcs := RunConfigurationSpec{}
			hash1 := rcs.ComputeHash()

			rcs.PipelineName = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("Schedule should change the hash", func() {
			rcs := RunConfigurationSpec{}
			hash1 := rcs.ComputeHash()

			rcs.Schedule = "notempty"
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))
		})

		Specify("All RuntimeParameters keys should change the hash", func() {
			rcs := RunConfigurationSpec{}
			hash1 := rcs.ComputeHash()

			rcs.RuntimeParameters = map[string]string{
				"a": "",
			}
			hash2 := rcs.ComputeHash()

			Expect(hash1).NotTo(Equal(hash2))

			rcs.RuntimeParameters = map[string]string{
				"b": "notempty",
			}
			hash3 := rcs.ComputeHash()

			Expect(hash2).NotTo(Equal(hash3))
		})
	})

	var _ = Describe("ComputeVersion", func() {

		Specify("Should have the spec hash only", func() {
			Expect(RunConfigurationSpec{}.ComputeVersion()).To(MatchRegexp("^[a-z0-9]{6}$"))
		})
	})
})
