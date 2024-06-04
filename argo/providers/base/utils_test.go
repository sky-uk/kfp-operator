//go:build unit

package base

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/argo/common"
)

var _ = Context("cron parser", func() {
	_ = Describe("should accept standard cron schedules", func() {
		schedule, err := ParseCron("  a   b c   d e f ")
		Expect(err).NotTo(HaveOccurred())
		Expect(schedule.PrintGo()).To(Equal("a b c d e f"))
		Expect(schedule.PrintStandard()).To(Equal("b c d e f"))
	})

	_ = Describe("should accept go cron schedules", func() {
		schedule, err := ParseCron("   b c   d e f ")
		Expect(err).NotTo(HaveOccurred())
		Expect(schedule.PrintGo()).To(Equal("0 b c d e f"))
		Expect(schedule.PrintStandard()).To(Equal("b c d e f"))
	})

	_ = Describe("should not parse when fields are missing", func() {
		_, err := ParseCron("* * * *")
		Expect(err).To(HaveOccurred())
	})

	_ = Describe("should not parse when too many fields are present", func() {
		_, err := ParseCron("* * * * * * *")
		Expect(err).To(HaveOccurred())
	})
})

var _ = Context("ResourceNameFromNamespacedName", func() {
	_ = Describe("should return string separated with hyphens", func() {
		result, err := ResourceNameFromNamespacedName(common.NamespacedName{
			Namespace: "my-namespace",
			Name:      "my-name",
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal("my-namespace-my-name"))
	})

	_ = Describe("should return error when namespace only set", func() {
		_, err := ResourceNameFromNamespacedName(common.NamespacedName{
			Namespace: "my-namespace",
		})
		Expect(err).To(HaveOccurred())
	})
})
