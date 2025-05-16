//go:build unit

package apis

import (
	"errors"
	"fmt"
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

	DescribeTable("Find",
		func(input []int, expected *int, expectedFound bool) {
			value, found := Find(input, func(i int) bool {
				return i == 2
			})
			Expect(found).To(Equal(expectedFound))
			if expected != nil {
				Expect(value).ToNot(BeNil())
				Expect(*value).To(Equal(*expected))
			} else {
				Expect(value).To(BeNil())
			}
		},
		Entry("", []int{1, 2, 3}, ptr(2), true),
		Entry("", []int{2, 1, 3}, ptr(2), true),
		Entry("", []int{1, 2, 3, 2}, ptr(2), true),
		Entry("", []int{1, 3, 5}, nil, false),
		Entry("", []int{}, nil, false),
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

	DescribeTable("MapValues", func(imv map[string]int, expected map[string]string) {
		Expect(MapValues(imv, func(a int) string {
			return fmt.Sprintf("output-%d", a)
		})).To(BeComparableTo(expected, cmpopts.EquateEmpty()))
	},
		Entry("", map[string]int{}, map[string]string{}),
		Entry("", nil, map[string]string{}),
		Entry("", map[string]int{"key1": 1, "key2": 2, "key3": 3}, map[string]string{"key1": "output-1", "key2": "output-2", "key3": "output-3"}),
	)
})

func ptr[T any](t T) *T {
	return &t
}
