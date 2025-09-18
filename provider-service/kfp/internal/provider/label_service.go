package provider

import (
	"encoding/json"
	"fmt"
	"github.com/tidwall/sjson"
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

func (dls DefaultLabelService) InsertLabelsIntoParameters(jsonBytes []byte, labels []string) ([]byte, error) {
	for _, label := range labels {
		var err error
		jsonBytes, err = sjson.SetRawBytes(jsonBytes, fmt.Sprintf("%s.%s", dls.parameterJsonPath, label), dls.parameterDefaults)
		if err != nil {
			return nil, fmt.Errorf("failed to inject label `%s` as parameter: %w", label, err)
		}
	}

	return jsonBytes, nil
}
