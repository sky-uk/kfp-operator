//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Utils", func() {
	DescribeTable("sliceDiff", func(as, bs, expected []int) {
		Expect(sliceDiff(as, bs, func(a int, b int) bool {
			return a == b
		})).To(Equal(expected))
	},
		Entry("", []int{1, 2}, []int{1}, []int{2}),
		Entry("", []int{2, 1}, []int{1}, []int{2}),
		Entry("", []int{1}, []int{}, []int{1}),
		Entry("", []int{1}, []int{1}, nil),
		Entry("", []int{}, []int{1}, nil),
	)
})
