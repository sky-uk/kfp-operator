package resource

import sigsyaml "sigs.k8s.io/yaml"

// unmarshalBody decodes a provider request body. Workflows submit YAML, and
// JSON remains valid so an updated provider can serve the previous operator.
func unmarshalBody(body []byte, dest any) error {
	return sigsyaml.Unmarshal(body, dest)
}
