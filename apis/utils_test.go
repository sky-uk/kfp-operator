//go:build unit

package apis

import (
	"errors"
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

	DescribeTable("MapErr", func(as []int, expected []string, expectSuccess bool) {
		actual, err := MapErr(as, func(a int) (string, error) {
			if a%2 == 0 {
				return "", errors.New("an error")
			}
			return strconv.Itoa(a), nil
		})
		if expectSuccess {
			Expect(err).To(BeNil())
			Expect(actual).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
		} else {
			Expect(err).NotTo(BeNil())
		}
	},
		Entry("", []int{}, []string{}, true),
		Entry("", []int{1, 3}, []string{"1", "3"}, true),
		Entry("", []int{3, 1}, []string{"3", "1"}, true),
		Entry("", []int{1, 2}, nil, false),
		Entry("", []int{2, 1}, nil, false),
	)

	type TestMapping struct {
		Key   string
		Value int
	}
})

func ptr[T any](t T) *T {
	return &t
}
