package provider

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"google.golang.org/protobuf/types/known/structpb"
	"regexp"
	"strings"
)

var (
	SchemaVersionNotFound = errors.New("expected 'schemaVersion' or 'pipelineSpec' in the pipeline values")
	SemVarV2              = semver.MustParse("2.0")
)

type PipelineValues struct {
	name         string
	labels       map[string]string
	pipelineSpec *structpb.Struct
}

type PipelineSchemaHandler interface {
	extract(raw map[string]any) (*PipelineValues, error)
}

type DefaultPipelineSchemaHandler struct {
	schema2Handler   PipelineSchemaHandler
	schema2_1Handler PipelineSchemaHandler
}

type Schema2Handler struct{}
type Schema2_1Handler struct{}

func extractSchemaVersion(raw map[string]any) (*semver.Version, error) {
	// 2.1 location of schemaVersion
	schemaVersion, ok := raw["schemaVersion"].(string)
	if !ok {
		// 2.0 location of schemaVersion
		pipelineSpec, ok := raw["pipelineSpec"].(map[string]any)
		if !ok {
			return nil, SchemaVersionNotFound
		}
		schemaVersion, ok = pipelineSpec["schemaVersion"].(string)
		if !ok {
			return nil, SchemaVersionNotFound
		}
	}
	version, err := semver.NewVersion(schemaVersion)
	if err != nil {
		return nil, err
	}
	return version, nil
}

func (dps DefaultPipelineSchemaHandler) extract(raw map[string]any) (*PipelineValues, error) {
	version, err := extractSchemaVersion(raw)

	if err != nil {
		return nil, err
	}
	if version.GreaterThan(SemVarV2) {
		return dps.schema2_1Handler.extract(raw)
	} else {
		return dps.schema2Handler.extract(raw)
	}
}

func (sv2 Schema2Handler) extract(raw map[string]any) (*PipelineValues, error) {
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

	labels, ok := raw["labels"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"expected map for 'labels', got %T",
			raw["labels"],
		)
	}
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

func (sv21 Schema2_1Handler) extract(raw map[string]any) (*PipelineValues, error) {
	pipelineInfo, ok := raw["pipelineInfo"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"expected map for 'pipelineInfo', got %T",
			raw["pipelineInfo"],
		)
	}
	displayName, ok := pipelineInfo["name"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'pipelineInfo.name', got %T",
			pipelineInfo["name"],
		)
	}

	schemaVersion, ok := raw["schemaVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'schemaVersion', got %T",
			raw["schemaVersion"],
		)
	}
	sdkVersion, ok := raw["sdkVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'sdkVersion', got %T",
			raw["sdkVersion"],
		)
	}
	labels := make(map[string]string)
	labels["schema_version"] = sanitizeString(schemaVersion)
	labels["sdk_version"] = sanitizeString(sdkVersion)

	pipelineSpecStruct, err := structpb.NewStruct(raw)
	if err != nil {
		return nil, err
	}

	return &PipelineValues{
		name:         displayName,
		labels:       labels,
		pipelineSpec: pipelineSpecStruct,
	}, nil
}

func sanitizeString(input string) string {
	input = strings.ToLower(input)
	re := regexp.MustCompile(`[^a-z0-9\-_]`)
	return re.ReplaceAllString(input, "_")
}
