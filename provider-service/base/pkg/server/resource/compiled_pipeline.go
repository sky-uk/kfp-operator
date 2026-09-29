package resource

import (
	"bytes"
	"encoding/json"
	"fmt"

	sigsyaml "sigs.k8s.io/yaml"
)

// CompiledPipeline is the opaque compiler output submitted to a provider.
// A YAML string preserves multiple documents, including a KFP platformSpec.
// A JSON object is accepted for one release and normalised to YAML.
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

	asYAML, err := sigsyaml.JSONToYAML(data)
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
