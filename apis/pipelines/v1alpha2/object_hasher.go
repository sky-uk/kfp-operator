package v1alpha2

import (
	"crypto/sha1"
	"hash"
	"sort"
)

var hashFieldSeparator = []byte{0}

//+kubebuilder:object:generate=false
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

func (oh ObjectHasher) WriteNamedValueListField(namedValues []NamedValue) {

	sort.Slice(namedValues, func(i, j int) bool {
		if namedValues[i].Name != namedValues[j].Name {
			return namedValues[i].Name < namedValues[j].Name
		} else {
			return namedValues[i].Value < namedValues[j].Value
		}
	})

	for _, k := range namedValues {
		oh.WriteStringField(k.Name)
		oh.WriteStringField(k.Value)
	}

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

func (oh ObjectHasher) Sum() []byte {
	return oh.h.Sum(nil)
}
