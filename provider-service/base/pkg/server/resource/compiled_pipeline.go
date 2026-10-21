package resource

import (
	"bytes"
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
)

// CompiledPipeline is an opaque compiler output submitted to a provider.
// It is represented as a byte slice containing the pipeline in YAML form.
//
// When decoding from JSON, CompiledPipeline accepts either:
//  - a YAML document encoded as a JSON string (quoted): the string contents
//    are used verbatim, preserving multi-document YAML such as a KFP platformSpec;
//  - a JSON object: the object is converted to YAML and stored.
//
// MarshalJSON encodes the stored YAML as a JSON string.
type CompiledPipeline []byte

func (c *CompiledPipeline) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)
	if len(data) == 0 || bytes.Equal(data, []byte("null")) {
		*c = nil
		return nil
	}

	if data[0] == '"' {
		var compiled string
		if err := json.Unmarshal(data, &compiled); err != nil {
			return err
		}
		*c = CompiledPipeline(compiled)
		return nil
	}

	asYAML, err := yaml.JSONToYAML(data)
	if err != nil {
		return fmt.Errorf("compiledPipeline must be a YAML string or JSON object: %w", err)
	}
	*c = CompiledPipeline(asYAML)
	return nil
}

func (c CompiledPipeline) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("null"), nil
	}
	return json.Marshal(string(c))
}
