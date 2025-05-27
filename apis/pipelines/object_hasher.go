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

func canonicalJSON(raw []byte) (string, error) {
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(canonical), nil
}

func (oh ObjectHasher) WriteJSONMapField(m map[string]*apiextensionsv1.JSON) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Validate all JSON before writing anything
	type pair struct {
		key       string
		canonical string
	}
	pairs := make([]pair, 0, len(m))

	for _, k := range keys {
		val := m[k]
		if val == nil {
			pairs = append(pairs, pair{key: k})
			continue
		}
		canonical, err := canonicalJSON(val.Raw)
		if err != nil {
			return // early exit, nothing written
		}
		pairs = append(pairs, pair{key: k, canonical: canonical})
	}

	// Only write to the hasher once all validation passes
	for _, p := range pairs {
		oh.WriteStringField(p.key)
		if p.canonical != "" {
			oh.WriteStringField(p.canonical)
		}
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
