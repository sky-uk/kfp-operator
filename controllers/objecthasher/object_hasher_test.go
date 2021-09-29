//go:build unit
// +build unit

package objecthasher

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ObjectHasher", func() {

	Specify("Adjacent string fields should be considered separate", func() {
		oh1 := New()
		oh1.WriteStringField("ab")
		oh1.WriteStringField("c")

		oh2 := New()
		oh2.WriteStringField("a")
		oh2.WriteStringField("bc")

		Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
	})

	Specify("Adjacent map fields should be considered separate", func() {
		oh1 := New()
		oh1.WriteMapField(map[string]string{
			"a": "1",
			"b": "2",
		})
		oh1.WriteMapField(map[string]string{
			"c": "3",
		})

		oh2 := New()
		oh2.WriteMapField(map[string]string{
			"a": "1",
		})
		oh2.WriteMapField(map[string]string{
			"b": "2",
			"c": "3",
		})

		Expect(oh1.Sum()).NotTo(Equal(oh2.Sum()))
	})

	Specify("Map field hash should be consistent", func() {
		iterations := 10 // simple: sometimes iterators are consistent
		sameMap := map[string]string{
			"a": "1",
			"b": "2",
		}

		for i := 0; i < iterations; i++ {
			oh1 := New()
			oh1.WriteMapField(sameMap)

			oh2 := New()
			oh2.WriteMapField(sameMap)

			Expect(oh1.Sum()).To(Equal(oh2.Sum()))
		}
	})
})
