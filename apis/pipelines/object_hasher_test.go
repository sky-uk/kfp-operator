//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/sky-uk/kfp-operator/apis"
)

var _ = Context("ObjectHasher", func() {
	var _ = Describe("WriteStringField", func() {
		Specify("Adjacent string fields should be considered separate", func() {
			oh1 := NewObjectHasher()
			oh1.WriteStringField("ab")
			oh1.WriteStringField("c")

			oh2 := NewObjectHasher()
			oh2.WriteStringField("a")
			oh2.WriteStringField("bc")

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})
	})

	var _ = Describe("WriteKVListField", func() {
		Specify("Adjacent NamedValue list fields should be considered separate", func() {
			oh1 := NewObjectHasher()
			WriteKVListField(oh1, []NamedValue{
				{Name: "a", Value: "1"},
				{Name: "b", Value: "2"},
			})
			WriteKVListField(oh1, []NamedValue{
				{Name: "c", Value: "3"},
			})

			oh2 := NewObjectHasher()
			WriteKVListField(oh2, []NamedValue{
				{Name: "a", Value: "1"},
			})
			WriteKVListField(oh2, []NamedValue{
				{Name: "b", Value: "2"},
				{Name: "c", Value: "3"},
			})

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("NamedValue list key and values should be considered separate", func() {
			oh1 := NewObjectHasher()
			WriteKVListField(oh1, []NamedValue{
				{Name: "ab", Value: "c"},
			})

			oh2 := NewObjectHasher()
			WriteKVListField(oh2, []NamedValue{
				{Name: "a", Value: "bc"},
			})

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("NamedValue list fields should be considered separate", func() {
			oh1 := NewObjectHasher()
			WriteKVListField(oh1, []NamedValue{
				{Name: "a", Value: "bc"},
				{Name: "d", Value: "e"},
			})

			oh2 := NewObjectHasher()
			WriteKVListField(oh2, []NamedValue{
				{Name: "a", Value: "b"},
				{Name: "cd", Value: "e"},
			})

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("NamedValue list field hash should be consistent if the order of entries is changed", func() {
			oh1 := NewObjectHasher()
			WriteKVListField(oh1, []NamedValue{
				{Name: "a", Value: "1"},
				{Name: "b", Value: "2"},
			})

			oh2 := NewObjectHasher()
			WriteKVListField(oh2, []NamedValue{
				{Name: "b", Value: "2"},
				{Name: "a", Value: "1"},
			})

			Expect(oh1.Sum()).To(Equal(oh2.Sum()))
		})

		Specify("NamedValue list field hash should be consistent if the order of multi-entries is changed", func() {
			oh1 := NewObjectHasher()
			WriteKVListField(oh1, []NamedValue{
				{Name: "a", Value: "1"},
				{Name: "a", Value: "2"},
			})

			oh2 := NewObjectHasher()
			WriteKVListField(oh2, []NamedValue{
				{Name: "a", Value: "2"},
				{Name: "a", Value: "1"},
			})

			Expect(oh1.Sum()).To(Equal(oh2.Sum()))
		})

		Specify("The original array should not be altered", func() {
			oh1 := NewObjectHasher()
			namedValues := []NamedValue{
				{Name: "b", Value: "1"},
				{Name: "a", Value: "2"},
			}

			WriteKVListField(oh1, namedValues)

			Expect(namedValues).To(Equal([]NamedValue{
				{Name: "b", Value: "1"},
				{Name: "a", Value: "2"},
			}))
		})
	})

	var _ = Describe("WriteMapField", func() {

		Specify("Adjacent map fields should be considered separate", func() {
			oh1 := NewObjectHasher()
			oh1.WriteMapField(map[string]string{
				"a": "1",
				"b": "2",
			})
			oh1.WriteMapField(map[string]string{
				"c": "3",
			})

			oh2 := NewObjectHasher()
			oh2.WriteMapField(map[string]string{
				"a": "1",
			})
			oh2.WriteMapField(map[string]string{
				"b": "2",
				"c": "3",
			})

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("Map key and values should be considered separate", func() {
			oh1 := NewObjectHasher()
			oh1.WriteMapField(map[string]string{
				"ab": "c",
			})

			oh2 := NewObjectHasher()
			oh2.WriteMapField(map[string]string{
				"a": "bc",
			})

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("Map fields should be considered separate", func() {
			oh1 := NewObjectHasher()
			oh1.WriteMapField(map[string]string{
				"a": "bc",
				"d": "e",
			})

			oh2 := NewObjectHasher()
			oh2.WriteMapField(map[string]string{
				"a":  "b",
				"cd": "e",
			})

			Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
		})

		Specify("Map field hash should be consistent", func() {
			// Map iterators start from a random point. The chance of
			// a false positive is {map len}^-{iterations}
			iterations := 10
			sameMap := map[string]string{
				"a": "1",
				"b": "2",
			}

			for i := 0; i < iterations; i++ {
				oh1 := NewObjectHasher()
				oh1.WriteMapField(sameMap)

				oh2 := NewObjectHasher()
				oh2.WriteMapField(sameMap)

				Expect(oh1.Sum()).To(Equal(oh2.Sum()))
			}
		})
	})
})
