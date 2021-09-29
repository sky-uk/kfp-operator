package objecthasher

import (
	"crypto/sha1"
	"hash"
	"sort"
)

var hashFieldSeparator = []byte{0}

type ObjectHasher struct {
	h hash.Hash
}

func New() ObjectHasher {
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
		oh.h.Write([]byte(k))
		oh.h.Write([]byte(value[k]))
	}

	oh.h.Write(hashFieldSeparator)
}

func (oh ObjectHasher) Sum() []byte {
	return oh.h.Sum(nil)
}
