//go:build unit
// +build unit

package v1alpha2

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Context("ObjectHasher", func() {
	Specify("Adjacent string fields should be considered separate", func() {
		oh1 := NewObjectHasher()
		oh1.WriteStringField("ab")
		oh1.WriteStringField("c")

		oh2 := NewObjectHasher()
		oh2.WriteStringField("a")
		oh2.WriteStringField("bc")

		Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
	})

	Specify("Adjacent NamedValue list fields should be considered separate", func() {
		oh1 := NewObjectHasher()
		oh1.WriteNamedValueListField([]NamedValue{
			{Name: "a", Value: "1"},
			{Name: "b", Value: "2"},
		})
		oh1.WriteNamedValueListField([]NamedValue{
			{Name: "c", Value: "3"},
		})

		oh2 := NewObjectHasher()
		oh2.WriteNamedValueListField([]NamedValue{
			{Name: "a", Value: "1"},
		})
		oh2.WriteNamedValueListField([]NamedValue{
			{Name: "b", Value: "2"},
			{Name: "c", Value: "3"},
		})

		Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
	})

	Specify("NamedValue list key and values should be considered separate", func() {
		oh1 := NewObjectHasher()
		oh1.WriteNamedValueListField([]NamedValue{
			{Name: "ab", Value: "c"},
		})

		oh2 := NewObjectHasher()
		oh2.WriteNamedValueListField([]NamedValue{
			{Name: "a", Value: "bc"},
		})

		Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
	})

	Specify("NamedValue list fields should be considered separate", func() {
		oh1 := NewObjectHasher()
		oh1.WriteNamedValueListField([]NamedValue{
			{Name: "a", Value: "bc"},
			{Name: "d", Value: "e"},
		})

		oh2 := NewObjectHasher()
		oh2.WriteNamedValueListField([]NamedValue{
			{Name: "a", Value: "b"},
			{Name: "cd", Value: "e"},
		})

		Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
	})

	Specify("NamedValue list field hash should be consistent if the order of entries is changed", func() {
		oh1 := NewObjectHasher()
		oh1.WriteNamedValueListField([]NamedValue{
			{Name: "a", Value: "1"},
			{Name: "b", Value: "2"},
		})

		oh2 := NewObjectHasher()
		oh2.WriteNamedValueListField([]NamedValue{
			{Name: "b", Value: "2"},
			{Name: "a", Value: "1"},
		})

		Expect(oh1.Sum()).To(Equal(oh2.Sum()))
	})

	Specify("NamedValue list field hash should be consistent if the order of multi-entries is changed", func() {
		oh1 := NewObjectHasher()
		oh1.WriteNamedValueListField([]NamedValue{
			{Name: "a", Value: "1"},
			{Name: "a", Value: "2"},
		})

		oh2 := NewObjectHasher()
		oh2.WriteNamedValueListField([]NamedValue{
			{Name: "a", Value: "2"},
			{Name: "a", Value: "1"},
		})

		Expect(oh1.Sum()).To(Equal(oh2.Sum()))
	})
})
