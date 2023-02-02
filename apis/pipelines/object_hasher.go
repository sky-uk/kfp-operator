package pipelines

import (
	"crypto/sha1"
	. "github.com/sky-uk/kfp-operator/apis"
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
	oh.h.Write(hashFieldSeparator)
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

	oh.h.Write(hashFieldSeparator)
}

func (oh ObjectHasher) WriteNamedValueListField(namedValues []NamedValue) {
	sorted := make([]NamedValue, len(namedValues))
	copy(sorted, namedValues)

	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Name != sorted[j].Name {
			return sorted[i].Name < sorted[j].Name
		} else {
			return sorted[i].Value < sorted[j].Value
		}
	})

	for _, k := range sorted {
		oh.WriteStringField(k.Name)
		oh.WriteStringField(k.Value)
	}

	oh.h.Write(hashFieldSeparator)
}

func (oh ObjectHasher) Sum() []byte {
	return oh.h.Sum(nil)
}
