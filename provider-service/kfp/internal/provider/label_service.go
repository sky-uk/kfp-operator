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

// InsertLabelsIntoParameters adds each label as a parameter of the PipelineSpec
// at the configured path (root.inputDefinitions.parameters by default),
// replacing any existing entry with the same name. Missing intermediate maps
// are created. It returns an error if the pipeline is empty or the path exists
// but is not a mapping.
//
// The YAML is edited as a node tree rather than decoded into Go values, so
// everything outside the inserted entries keeps its original text and quoting.
// A trailing PlatformSpec document is preserved.
func (dls DefaultLabelService) InsertLabelsIntoParameters(
	pipeline resource.CompiledPipeline,
	labels []string,
) (resource.CompiledPipeline, error) {
	documents, err := decodeYAMLDocuments([]byte(pipeline))
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, fmt.Errorf("compiled pipeline is empty")
	}

	// The defaults are plain scalars, so encoding them is safe. Encoding
	// arbitrary values via Node.Encode is not: it round-trips through the
	// emitter described on encodeYAMLDocuments.
	var defaults yamlv3.Node
	if err := defaults.Encode(dls.parameterDefaults); err != nil {
		return nil, fmt.Errorf("failed to encode label parameter defaults: %w", err)
	}

	parameters, err := getOrCreateNestedMapping(documents[0].Content[0], dls.parameterPath...)
	if err != nil {
		return nil, fmt.Errorf("failed to access pipeline parameters: %w", err)
	}

	for _, label := range labels {
		setMappingEntry(parameters, label, &defaults)
	}

	encoded, err := encodeYAMLDocuments(documents)
	if err != nil {
		return nil, err
	}
	return resource.CompiledPipeline(encoded), nil
}

// decodeYAMLDocuments decodes a multi-document YAML byte slice into one node
// tree per document. A KFP IR YAML contains a PipelineSpec and an optional
// PlatformSpec as separate documents. Empty documents are skipped.
func decodeYAMLDocuments(pipeline []byte) ([]*yamlv3.Node, error) {
	decoder := yamlv3.NewDecoder(bytes.NewReader(pipeline))
	var documents []*yamlv3.Node
	for {
		var document yamlv3.Node
		err := decoder.Decode(&document)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to decode compiled pipeline: %w", err)
		}
		if document.Kind == 0 || len(document.Content) == 0 {
			continue
		}
		documents = append(documents, &document)
	}
	return documents, nil
}

// getOrCreateNestedMapping traverses the mapping node root along path,
// creating intermediate mappings as needed, and returns the innermost mapping.
func getOrCreateNestedMapping(root *yamlv3.Node, path ...string) (*yamlv3.Node, error) {
	if root.Kind != yamlv3.MappingNode {
		return nil, fmt.Errorf("expected a mapping at the document root, got %s", root.Tag)
	}

	current := root
	for _, key := range path {
		next := mappingValue(current, key)
		if next == nil {
			next = &yamlv3.Node{Kind: yamlv3.MappingNode, Tag: "!!map"}
			setMappingEntry(current, key, next)
		} else if next.Kind != yamlv3.MappingNode {
			return nil, fmt.Errorf("expected map at key %q, got %s", key, next.Tag)
		}
		current = next
	}
	return current, nil
}

// mappingValue returns the value stored under key in the mapping node m, or
// nil if the key is absent.
func mappingValue(m *yamlv3.Node, key string) *yamlv3.Node {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return m.Content[i+1]
		}
	}
	return nil
}

// setMappingEntry sets key to value in the mapping node m, replacing an
// existing entry or appending a new one.
func setMappingEntry(m *yamlv3.Node, key string, value *yamlv3.Node) {
	if existing := mappingValue(m, key); existing != nil {
		*existing = *value
		return
	}
	keyNode := &yamlv3.Node{Kind: yamlv3.ScalarNode, Tag: "!!str", Value: key}
	m.Content = append(m.Content, keyNode, value)
}

// encodeYAMLDocuments encodes document node trees into a single multi-document
// YAML byte slice, separated by --- markers.
//
// The encoder must indent by 2. go.yaml.in/yaml/v3 writes the block scalar
// indentation indicator (e.g. "|4") from the encoder's indent width, but
// indents list item content by only 2. With any other width, block scalars
// that begin with a blank line or whitespace get an indicator that does not
// match their content and cannot be parsed back. Some KFP API server versions
// (e.g. 2.14.4) hang on such input. The KFP SDK emits these scalars in every
// executor command.
//
// TODO: revisit once a release of go.yaml.in/yaml contains the fix for
// https://github.com/yaml/go-yaml/issues/76 (fixed on main by
// https://github.com/yaml/go-yaml/pull/414, not yet in a v3 release).
func encodeYAMLDocuments(documents []*yamlv3.Node) ([]byte, error) {
	var encoded bytes.Buffer
	encoder := yamlv3.NewEncoder(&encoded)
	encoder.SetIndent(2)
	for _, document := range documents {
		if err := encoder.Encode(document); err != nil {
			return nil, fmt.Errorf("failed to encode compiled pipeline: %w", err)
		}
	}
	if err := encoder.Close(); err != nil {
		return nil, fmt.Errorf("failed to encode compiled pipeline: %w", err)
	}
	return encoded.Bytes(), nil
}
