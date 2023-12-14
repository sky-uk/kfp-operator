//go:build unit

package pipelines

import (
	"errors"
	"github.com/google/go-cmp/cmp/cmpopts"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"strconv"
	"strings"
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

	DescribeTable("Exists", func(as []int, expected bool) {
		Expect(Exists(as, func(a int) bool {
			return a%2 == 0
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{}, false),
		Entry("", []int{2}, true),
		Entry("", []int{1}, false),
		Entry("", []int{1, 2}, true),
		Entry("", []int{2, 1}, true),
		Entry("", []int{1, 3}, false),
	)

	DescribeTable("Contains", func(as []int, expected bool) {
		Expect(Contains(as, 2)).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{}, false),
		Entry("", []int{2}, true),
		Entry("", []int{1}, false),
		Entry("", []int{1, 2}, true),
		Entry("", []int{2, 1}, true),
	)

	DescribeTable("Forall", func(as []int, expected bool) {
		Expect(Forall(as, func(a int) bool {
			return a%2 == 0
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{}, true),
		Entry("", []int{1}, false),
		Entry("", []int{2}, true),
		Entry("", []int{1, 2}, false),
		Entry("", []int{2, 1}, false),
		Entry("", []int{2, 4}, true),
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

	DescribeTable("Flatten", func(as [][]int, expected []int) {
		Expect(Flatten(as...)).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", [][]int{}, []int{}),
		Entry("", [][]int{{}, {}}, []int{}),
		Entry("", [][]int{{1, 2}, {3, 4}}, []int{1, 2, 3, 4}),
	)

	DescribeTable("FlatMap", func(as []string, expected []string) {
		Expect(FlatMap(as, func(a string) []string {
			return strings.Split(a, " ")
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []string{}, []string{}),
		Entry("", []string{"1 2", "3 4"}, []string{"1", "2", "3", "4"}),
		Entry("", []string{"3 4", "1 2"}, []string{"3", "4", "1", "2"}),
	)

	DescribeTable("FlatMapErr", func(as []string, expected []string, expectSuccess bool) {
		actual, err := FlatMapErr(as, func(a string) ([]string, error) {
			if a == "err" {
				return nil, errors.New("an error")
			}
			return strings.Split(a, " "), nil
		})

		if expectSuccess {
			Expect(err).To(BeNil())
			Expect(actual).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
		} else {
			Expect(err).NotTo(BeNil())
		}
	},
		Entry("", []string{}, []string{}, true),
		Entry("", []string{"1 2", "3 4"}, []string{"1", "2", "3", "4"}, true),
		Entry("", []string{"1 2", "err"}, nil, false),
		Entry("", []string{"3 4", "1 2"}, []string{"3", "4", "1", "2"}, true),
		Entry("", []string{"err", "1 2"}, nil, false),
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

	DescribeTable("GroupMap", func(as []int, expected map[string][]int) {
		Expect(GroupMap(as, func(a int) (string, int) {
			return strconv.Itoa(a / 10), a % 10
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{11, 12, 23}, map[string][]int{"1": {1, 2}, "2": {3}}),
		Entry("", []int{23, 11, 12}, map[string][]int{"1": {1, 2}, "2": {3}}),
		Entry("", []int{11, 12}, map[string][]int{"1": {1, 2}}),
		Entry("", []int{}, map[string][]int{}),
	)

	DescribeTable("Duplicates", func(as []int, expected []int) {
		Expect(Duplicates(as)).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{1, 1, 2}, []int{1}),
		Entry("", []int{1, 2, 2}, []int{2}),
		Entry("", []int{1, 1, 1}, []int{1}),
		Entry("", []int{1, 2, 3}, []int{}),
		Entry("", []int{}, []int{}),
	)

	DescribeTable("Unique", func(as []int, expected []int) {
		Expect(Unique(as)).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []int{}, []int{}),
		Entry("", []int{1, 1, 2}, []int{1, 2}),
		Entry("", []int{1, 2, 2}, []int{1, 2}),
		Entry("", []int{1, 1, 1}, []int{1}),
		Entry("", []int{1, 2, 3}, []int{1, 2, 3}),
	)

	type TestMapping struct {
		Key   string
		Value int
	}
	DescribeTable("ToMap", func(kvs []TestMapping, expected map[string]int) {
		Expect(ToMap(kvs, func(a TestMapping) (string, int) {
			return a.Key, a.Value
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", []TestMapping{}, map[string]int{}),
		Entry("", []TestMapping{{"1", 1}, {"2", 2}}, map[string]int{"1": 1, "2": 2}),
		Entry("", []TestMapping{{"1", 1}, {"1", 3}, {"2", 2}}, map[string]int{"1": 3, "2": 2}),
	)

	DescribeTable("Values", func(kvs map[string]int, expected []int) {
		Expect(Values(kvs)).To(ConsistOf(expected))
	},
		Entry("", map[string]int{}, []int{}),
		Entry("", map[string]int{"1": 1, "2": 2}, []int{1, 2}),
		Entry("", map[string]int{"1": 1, "2": 1}, []int{1, 1}),
	)
})
