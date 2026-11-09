package provider

import (
	"fmt"
	"sort"
	"strings"

	yamlv3 "go.yaml.in/yaml/v3"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

// Hard-coded ConfigMap settings applied to TFX executor pods. The pod volume is
// named after the ConfigMap, so hackConfigMapName must be a lowercase DNS-1123
// label.
const (
	// Mounted as a volume.
	hackConfigMapName = "foo-config"
	hackMountPath     = "/etc/placeholder"

	// Exposed as an env var. The ConfigMap and key must exist, otherwise the pod
	// fails with CreateContainerConfigError.
	hackEnvConfigMapName = "bar-env-config"
	hackEnvConfigMapKey  = "some-key"
	hackEnvVarName       = "SOME_ENV_VAR"
)

// addHackedTFXPlatformSpec replaces the PlatformSpec document of a TFX pipeline
// with one that, on every container executor, mounts hackConfigMapName at
// hackMountPath and sets hackEnvVarName from hackEnvConfigMapKey of
// hackEnvConfigMapName. Other frameworks are returned unchanged.
func addHackedTFXPlatformSpec(
	pipeline resource.CompiledPipeline,
	frameworkName string,
) (resource.CompiledPipeline, error) {
	if strings.ToLower(frameworkName) != "tfx" {
		return pipeline, nil
	}

	documents, err := decodeYAMLDocuments([]byte(pipeline))
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, fmt.Errorf("compiled pipeline is empty")
	}

	executorNames, err := containerExecutorNames(documents[0].Content[0])
	if err != nil {
		return nil, err
	}
	if len(executorNames) == 0 {
		// No container executors.
		return pipeline, nil
	}

	executors := map[string]any{}
	for _, name := range executorNames {
		executors[name] = map[string]any{
			"configMapAsVolume": []any{
				map[string]any{
					"configMapName": hackConfigMapName,
					"mountPath":     hackMountPath,
					"optional":      false,
				},
			},
			"configMapAsEnv": []any{
				map[string]any{
					"configMapName": hackEnvConfigMapName,
					"keyToEnv": []any{
						map[string]any{
							"configMapKey": hackEnvConfigMapKey,
							"envVar":       hackEnvVarName,
						},
					},
				},
			},
		}
	}

	var platformSpec yamlv3.Node
	if err := platformSpec.Encode(map[string]any{
		"platforms": map[string]any{
			"kubernetes": map[string]any{
				"deploymentSpec": map[string]any{
					"executors": executors,
				},
			},
		},
	}); err != nil {
		return nil, fmt.Errorf("failed to encode platform spec: %w", err)
	}

	// Replaces any existing PlatformSpec.
	encoded, err := encodeYAMLDocuments([]*yamlv3.Node{documents[0], &platformSpec})
	if err != nil {
		return nil, err
	}
	return resource.CompiledPipeline(encoded), nil
}

// containerExecutorNames returns the sorted names of the executors that run a
// container, excluding importer and resolver executors.
func containerExecutorNames(pipelineSpec *yamlv3.Node) ([]string, error) {
	if pipelineSpec.Kind != yamlv3.MappingNode {
		return nil, fmt.Errorf("expected a mapping at the pipeline spec root, got %s", pipelineSpec.Tag)
	}

	deploymentSpec := mappingValue(pipelineSpec, "deploymentSpec")
	if deploymentSpec == nil {
		return nil, nil
	}
	executors := mappingValue(deploymentSpec, "executors")
	if executors == nil {
		return nil, nil
	}
	if executors.Kind != yamlv3.MappingNode {
		return nil, fmt.Errorf("expected a mapping at deploymentSpec.executors, got %s", executors.Tag)
	}

	var names []string
	for i := 0; i+1 < len(executors.Content); i += 2 {
		executor := executors.Content[i+1]
		if executor.Kind == yamlv3.MappingNode && mappingValue(executor, "container") != nil {
			names = append(names, executors.Content[i].Value)
		}
	}
	sort.Strings(names)
	return names, nil
}
