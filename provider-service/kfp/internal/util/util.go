package util

import "fmt"

// ByDisplayNameFilter creates a url-encoded, JSON-serialized Filter protocol buffer.
// https://github.com/kubeflow/pipelines/blob/2.1.0/backend/api/v2beta1/filter.proto
func ByDisplayNameFilter(name string) string {
	return fmt.Sprintf(
		`{"predicates": [{"operation": "EQUALS", "key": "display_name", "string_value": "%s"}]}`,
		name,
	)
}
