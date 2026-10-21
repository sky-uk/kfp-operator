package provider

import (
	"bytes"
	"fmt"
	"io"

	yamlv3 "go.yaml.in/yaml/v3"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

// LabelService inserts tracking labels into a compiled pipeline's parameters.
type LabelService interface {
	InsertLabelsIntoParameters(
		pipeline resource.CompiledPipeline,
		labels []string,
	) (resource.CompiledPipeline, error)
}

// DefaultLabelService is the default LabelService implementation, inserting
// labels as optional string parameters at a configurable YAML path.
type DefaultLabelService struct {
	parameterDefaults map[string]any
	parameterPath     []string
}

// NewDefaultLabelService returns a DefaultLabelService configured to insert
// labels into root.inputDefinitions.parameters as optional STRING parameters.
func NewDefaultLabelService() LabelService {
	return &DefaultLabelService{
		parameterDefaults: map[string]any{
			"isOptional":    true,
			"parameterType": "STRING",
			"defaultValue":  "",
		},
		parameterPath: []string{"root", "inputDefinitions", "parameters"},
	}
}

// InsertLabelsIntoParameters injects each label as an optional string parameter
// into root.inputDefinitions.parameters of the PipelineSpec. Any trailing
// PlatformSpec document is preserved unchanged.
func (dls DefaultLabelService) InsertLabelsIntoParameters(
	pipeline resource.CompiledPipeline,
	labels []string,
) (resource.CompiledPipeline, error) {
	documents, err := splitYAMLDocuments([]byte(pipeline))
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, fmt.Errorf("compiled pipeline is empty")
	}

	var spec map[string]any
	if err := yamlv3.Unmarshal(documents[0], &spec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PipelineSpec: %w", err)
	}

	parameters, err := getOrCreateNestedMap(spec, dls.parameterPath...)
	if err != nil {
		return nil, fmt.Errorf("failed to access pipeline parameters: %w", err)
	}

	for _, label := range labels {
		parameters[label] = dls.parameterDefaults
	}

	specYAML, err := yamlv3.Marshal(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to encode compiled pipeline: %w", err)
	}
	documents[0] = specYAML
	return resource.CompiledPipeline(joinYAMLDocuments(documents)), nil
}

// getOrCreateNestedMap traverses m along path, creating intermediate maps as
// needed, and returns the innermost map.
func getOrCreateNestedMap(m map[string]any, path ...string) (map[string]any, error) {
	current := m
	for _, key := range path {
		val, ok := current[key]
		if !ok {
			nested := map[string]any{}
			current[key] = nested
			current = nested
			continue
		}
		nested, ok := val.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected map at key %q, got %T", key, val)
		}
		current = nested
	}
	return current, nil
}

// splitYAMLDocuments splits a multi-document YAML byte slice into individual
// documents. A KFP IR YAML contains a PipelineSpec and an optional PlatformSpec
// as separate documents.
func splitYAMLDocuments(pipeline []byte) ([][]byte, error) {
	decoder := yamlv3.NewDecoder(bytes.NewReader(pipeline))
	var documents [][]byte
	for {
		var document yamlv3.Node
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode compiled pipeline: %w", err)
		}
		if document.Kind == 0 {
			continue
		}
		encoded, err := yamlv3.Marshal(&document)
		if err != nil {
			return nil, fmt.Errorf("failed to re-encode YAML document: %w", err)
		}
		documents = append(documents, encoded)
	}
	return documents, nil
}

// joinYAMLDocuments concatenates YAML documents into a single byte slice,
// separated by --- markers.
func joinYAMLDocuments(documents [][]byte) []byte {
	var encoded bytes.Buffer
	for i, document := range documents {
		if i > 0 {
			encoded.WriteString("---\n")
		}
		encoded.Write(bytes.TrimSpace(document))
		encoded.WriteByte('\n')
	}
	return encoded.Bytes()
}
