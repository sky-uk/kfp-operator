package provider

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/structpb"
)

// SchemaVersion2_1 is the schemaVersion at which Vertex AI's PipelineJob
// requires RuntimeConfig.ParameterValues instead of the deprecated Parameters.
var SchemaVersion2_1 = semver.MustParse("2.1.0")

type PipelineValues struct {
	name          string
	labels        map[string]string
	pipelineSpec  *structpb.Struct
	schemaVersion *semver.Version
}

// DefaultPipelineSchemaHandler handles both the bare KFP pipeline spec and the
// TFX PipelineJob wrapper, for either schemaVersion 2.0 or 2.1.
type DefaultPipelineSchemaHandler struct{}

func (DefaultPipelineSchemaHandler) extract(raw map[string]any) (*PipelineValues, error) {
	spec, name, wrapperLabels := unwrap(raw)

	// An empty name means the spec is not wrapped, so take the name from pipelineInfo.
	if name == "" {
		pipelineInfo, ok := spec["pipelineInfo"].(map[string]any)
		if !ok {
			return nil, fmt.Errorf(
				"expected map for 'pipelineInfo', got %T",
				spec["pipelineInfo"],
			)
		}
		name, ok = pipelineInfo["name"].(string)
		if !ok {
			return nil, fmt.Errorf(
				"expected string for 'pipelineInfo.name', got %T",
				pipelineInfo["name"],
			)
		}
	}

	schemaVersion, ok := spec["schemaVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'schemaVersion', got %T",
			spec["schemaVersion"],
		)
	}
	parsedSchemaVersion, err := semver.NewVersion(schemaVersion)
	if err != nil {
		return nil, fmt.Errorf("invalid 'schemaVersion' %q: %w", schemaVersion, err)
	}
	sdkVersion, ok := spec["sdkVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'sdkVersion', got %T",
			spec["sdkVersion"],
		)
	}

	labels := lo.MapValues(wrapperLabels, func(value any, _ string) string {
		if stringValue, ok := value.(string); ok {
			return stringValue
		}
		return fmt.Sprintf("%v", value)
	})
	labels["schema_version"] = schemaVersion
	labels["sdk_version"] = sdkVersion

	pipelineSpecStruct, err := structpb.NewStruct(spec)
	if err != nil {
		return nil, err
	}

	return &PipelineValues{
		name:          name,
		labels:        labels,
		pipelineSpec:  pipelineSpecStruct,
		schemaVersion: parsedSchemaVersion,
	}, nil
}

// unwrap returns the inner spec plus the wrapper's displayName and labels when
// raw is a TFX PipelineJob wrapper, otherwise raw itself with no wrapper values.
func unwrap(raw map[string]any) (spec map[string]any, displayName string, labels map[string]any) {
	pipelineSpec, isWrapped := raw["pipelineSpec"].(map[string]any)
	if !isWrapped {
		return raw, "", nil
	}
	displayName, _ = raw["displayName"].(string)
	labels, _ = raw["labels"].(map[string]any)
	return pipelineSpec, displayName, labels
}
