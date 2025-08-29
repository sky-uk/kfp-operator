package provider

import (
	"encoding/json"
	"fmt"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"google.golang.org/protobuf/types/known/structpb"
)

type PipelineValues struct {
	name         string
	labels       map[string]string
	pipelineSpec *structpb.Struct
}

type PipelineSchemaHandler interface {
	extract(compiledPipeline resource.CompiledPipeline) (*PipelineValues, error)
}

type SchemaHandler struct{}

func (sh SchemaHandler) extract(
	compiledPipeline resource.CompiledPipeline,
) (*PipelineValues, error) {
	var pipelineSpec map[string]any
	if err := json.Unmarshal(compiledPipeline.PipelineSpec, &pipelineSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PipelineSpec: %w", err)
	}

	pipelineSpecStruct, err := structpb.NewStruct(pipelineSpec)
	if err != nil {
		return nil, err
	}

	schemaVersion, ok := pipelineSpec["schemaVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'schemaVersion', got %T",
			pipelineSpec["schemaVersion"],
		)
	}
	sdkVersion, ok := pipelineSpec["sdkVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'sdkVersion', got %T",
			pipelineSpec["sdkVersion"],
		)
	}

	compiledPipeline.Labels["schema_version"] = schemaVersion
	compiledPipeline.Labels["sdk_version"] = sdkVersion

	return &PipelineValues{
		name:         compiledPipeline.DisplayName,
		labels:       compiledPipeline.Labels,
		pipelineSpec: pipelineSpecStruct,
	}, nil
}
