package pipelines

import (
	"crypto/sha1"
	"encoding/json"
	"hash"
	"sort"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// canonicalJSONOrRaw checks if the input raw bytes is valid JSON.
// If it is NOT valid JSON then return the stringified bytes
// If it IS valid JSON then return the canonical stringified bytes
func canonicalJSONOrRaw(raw []byte) string {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}

	canonical, err := json.Marshal(v)
	if err != nil {
		return string(raw)
	}

	return string(canonical)
}

// WriteJSONMapField hashes a map of string to raw bytes.
// If raw bytes is valid JSON then it will hash the canonical JSON form
// If raw bytes is invalid JSON then it will hash the value directly
// In practice this is redundant because K8s API server will guarantee fields
// defined as Raw bytes are valid JSON.
func (oh ObjectHasher) WriteJSONMapField(m map[string]*apiextensionsv1.JSON) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		value := m[k]
		oh.WriteStringField(k)

		if value == nil {
			continue
		}

		canonicalOrRaw := canonicalJSONOrRaw(value.Raw)
		oh.WriteStringField(canonicalOrRaw)
	}
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
