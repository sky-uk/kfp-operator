package provider

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"
)

type PipelineValues struct {
	name         string
	labels       map[string]string
	pipelineSpec *structpb.Struct
}

type PipelineSchemaHandler interface {
	extract(raw map[string]any) (*PipelineValues, error)
}

type SchemaHandler struct{}

func (sh SchemaHandler) extract(raw map[string]any) (*PipelineValues, error) {
	displayName, ok := raw["displayName"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'displayName', got %T",
			raw["displayName"],
		)
	}

	pipelineSpec, ok := raw["pipelineSpec"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"expected map for 'pipelineSpec', got %T",
			raw["pipelineSpec"],
		)
	}

	pipelineSpecStruct, err := structpb.NewStruct(pipelineSpec)
	if err != nil {
		return nil, err
	}

	schemaVersion, ok := pipelineSpec["schemaVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'schemaVersion', got %T",
			raw["schemaVersion"],
		)
	}
	sdkVersion, ok := pipelineSpec["sdkVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'sdkVersion', got %T",
			raw["sdkVersion"],
		)
	}

	labels, ok := raw["labels"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"expected map for 'labels', got %T",
			raw["labels"],
		)
	}
	labels["schema_version"] = schemaVersion
	labels["sdk_version"] = sdkVersion

	convertedLabels := make(map[string]string)
	for k, v := range labels {
		if strVal, ok := v.(string); ok {
			convertedLabels[k] = strVal
		} else {
			convertedLabels[k] = fmt.Sprintf("%v", v)
		}
	}

	return &PipelineValues{
		name:         displayName,
		labels:       convertedLabels,
		pipelineSpec: pipelineSpecStruct,
	}, nil
}
