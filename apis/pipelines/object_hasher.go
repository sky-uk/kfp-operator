package pipelines

import (
	"crypto/sha1"
	"hash"
	"sort"
)

var hashFieldSeparator = []byte{0}

type ObjectHasher struct {
	h hash.Hash
}

func NewObjectHasher() ObjectHasher {
	return ObjectHasher{
		sha1.New(),
	}
}

func (oh ObjectHasher) WriteStringField(value string) {
	oh.h.Write([]byte(value))
	oh.WriteFieldSeparator()
}

func (oh ObjectHasher) WriteMapField(value map[string]string) {
	keys := make([]string, 0, len(value))
	for k := range value {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		oh.WriteStringField(k)
		oh.WriteStringField(value[k])
	}

	oh.WriteFieldSeparator()
}

type KV interface {
	GetKey() string
	GetValue() string
}

// Has to be a function with parameter because Go does not support generic methods
func WriteKVListField[T KV](oh ObjectHasher, kvs []T) {
	sorted := make([]T, len(kvs))
	copy(sorted, kvs)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].GetKey() != sorted[j].GetKey() {
			return sorted[i].GetKey() < sorted[j].GetKey()
		} else {
			return sorted[i].GetValue() < sorted[j].GetValue()
		}
	})

	for _, k := range sorted {
		oh.WriteStringField(k.GetKey())
		oh.WriteStringField(k.GetValue())
	}

	oh.WriteFieldSeparator()
}

func (oh ObjectHasher) WriteFieldSeparator() {
	oh.h.Write(hashFieldSeparator)
}

func (oh ObjectHasher) Sum() []byte {
	return oh.h.Sum(nil)
}
