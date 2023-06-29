//go:build unit
// +build unit

package apis

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("Utils", func() {
	DescribeTable("sliceDiff", func(as, bs, expected []int) {
		Expect(SliceDiff(as, bs, func(a int, b int) bool {
			return a == b
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{1, 2}, []int{1}, []int{2}),
		Entry("", []int{2, 1}, []int{1}, []int{2}),
		Entry("", []int{1}, []int{}, []int{1}),
		Entry("", []int{1}, []int{1}, []int{}),
		Entry("", []int{}, []int{1}, []int{}),
	)

	DescribeTable("filter", func(as, expected []int) {
		Expect(Filter(as, func(a int) bool {
			return a%2 == 0
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{1, 2}, []int{2}),
		Entry("", []int{2, 1}, []int{2}),
		Entry("", []int{1}, []int{}),
		Entry("", []int{}, []int{}),
	)
})
