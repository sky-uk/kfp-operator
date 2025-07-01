package util

import "fmt"

// ByNameFilter creates a url-encoded, JSON-serialized Filter protocol buffer.
// https://github.com/kubeflow/pipelines/blob/master/backend/api/v1beta1/filter.proto
func ByNameFilter(name string) string {
	return fmt.Sprintf(
		`{"predicates": [{"op": "EQUALS", "key": "name", "string_value": "%s"}]}`,
		name,
	)
}
