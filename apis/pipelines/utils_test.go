//go:build unit
// +build unit

package pipelines

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strconv"
)

var _ = Context("Utils", func() {
	DescribeTable("SliceDiff", func(as, bs, expected []int) {
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

	DescribeTable("Filter", func(as, expected []int) {
		Expect(Filter(as, func(a int) bool {
			return a%2 == 0
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{1, 2}, []int{2}),
		Entry("", []int{2, 1}, []int{2}),
		Entry("", []int{1}, []int{}),
		Entry("", []int{}, []int{}),
	)

	DescribeTable("Map", func(as []int, expected []string) {
		Expect(Map(as, func(a int) string {
			return strconv.Itoa(a)
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{1, 2}, []string{"1", "2"}),
		Entry("", []int{2, 1}, []string{"2", "1"}),
		Entry("", []int{}, []string{}),
	)

	DescribeTable("Collect", func(as []int, expected []string) {
		Expect(Collect(as, func(a int) (string, bool) {
			if a%2 == 0 {
				return strconv.Itoa(a), true
			}

			return "", false
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{1, 2}, []string{"2"}),
		Entry("", []int{2, 1}, []string{"2"}),
		Entry("", []int{1}, []string{}),
		Entry("", []int{}, []string{}),
	)
})
