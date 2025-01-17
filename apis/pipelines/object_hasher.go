package pipelines

import (
	"crypto/sha1"
	"encoding/json"
	"hash"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (oh ObjectHasher) WriteObject(obj metav1.Object) error {
	bytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	oh.h.Write(bytes)
	return nil
}

type KV interface {
	GetKey() string
	GetValue() string
}

// Has to be a function with parameter because Go does not support generic methods
func WriteKVListField[T KV](oh ObjectHasher, kvs []T) {
	WriteList(oh, kvs, func(kv1, kv2 T) bool {
		if kv1.GetKey() != kv2.GetKey() {
			return kv1.GetKey() < kv2.GetKey()
		} else {
			return kv1.GetValue() < kv2.GetValue()
		}
	}, func(oh ObjectHasher, kv T) {
		oh.WriteStringField(kv.GetKey())
		oh.WriteStringField(kv.GetValue())
	})
}

// Has to be a function with parameter because Go does not support generic methods
func WriteList[T any](oh ObjectHasher, ts []T, cmp func(t1, t2 T) bool, write func(oh ObjectHasher, t T)) {
	sorted := make([]T, len(ts))
	copy(sorted, ts)

	sort.Slice(sorted, func(i, j int) bool {
		return cmp(sorted[i], sorted[j])
	})

	for _, k := range sorted {
		write(oh, k)
	}

	oh.WriteFieldSeparator()
}

func (oh ObjectHasher) WriteFieldSeparator() {
	oh.h.Write(hashFieldSeparator)
}

func (oh ObjectHasher) Sum() []byte {
	return oh.h.Sum(nil)
}
