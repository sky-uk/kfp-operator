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
