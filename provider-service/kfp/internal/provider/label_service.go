package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/tidwall/sjson"
	yamlv3 "go.yaml.in/yaml/v3"
	sigsyaml "sigs.k8s.io/yaml"
)

type LabelService interface {
	InsertLabelsIntoParameters(jsonBytes []byte, labels []string) ([]byte, error)
}

type DefaultLabelService struct {
	parameterDefaults []byte
	parameterJsonPath string
}

func NewDefaultLabelService() (LabelService, error) {
	paramDef := map[string]any{
		"isOptional":    true,
		"parameterType": "STRING",
		"defaultValue":  "",
	}

	paramAsJson, err := json.Marshal(paramDef)
	if err != nil {
		return nil, err
	}

	return &DefaultLabelService{
		parameterDefaults: paramAsJson,
		parameterJsonPath: "root.inputDefinitions.parameters",
	}, nil

}

func (dls DefaultLabelService) InsertLabelsIntoParameters(pipeline []byte, labels []string) ([]byte, error) {
	documents, err := splitYAMLDocuments(pipeline)
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, fmt.Errorf("compiled pipeline is empty")
	}

	specJSON, err := sigsyaml.YAMLToJSON(documents[0])
	if err != nil {
		return nil, fmt.Errorf("failed to read compiled pipeline: %w", err)
	}
	for _, label := range labels {
		specJSON, err = sjson.SetRawBytes(specJSON, fmt.Sprintf("%s.%s", dls.parameterJsonPath, label), dls.parameterDefaults)
		if err != nil {
			return nil, fmt.Errorf("failed to inject label `%s` as parameter: %w", label, err)
		}
	}

	specYAML, err := sigsyaml.JSONToYAML(specJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to encode compiled pipeline: %w", err)
	}
	documents[0] = specYAML
	return joinYAMLDocuments(documents), nil
}

// splitYAMLDocuments separates a compiled pipeline into its YAML documents.
// KFP writes pipelineSpec and platformSpec as two documents when platform features are used.
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
			return nil, fmt.Errorf("failed to read compiled pipeline: %w", err)
		}
		if document.Kind == 0 {
			continue
		}
		encoded, err := yamlv3.Marshal(&document)
		if err != nil {
			return nil, fmt.Errorf("failed to read compiled pipeline: %w", err)
		}
		documents = append(documents, encoded)
	}
	return documents, nil
}

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
