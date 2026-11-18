package provider

import (
	"fmt"
	"os"
	"sort"
	"strings"

	yamlv3 "go.yaml.in/yaml/v3"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

// Environment variables read from the provider service container. The volume
// pair and the env pair are each optional, but both variables of a pair must be
// set together.
const (
	// The ConfigMap is mounted as a volume named after it, so it must be a
	// lowercase DNS-1123 label.
	volumeConfigMapEnvVar = "TFX_EXECUTOR_VOLUME_CONFIGMAP"
	volumeMountPathEnvVar = "TFX_EXECUTOR_VOLUME_MOUNT_PATH"

	// The ConfigMap key is exposed as an env var of the same name. The ConfigMap
	// and key must exist, otherwise the pod fails with CreateContainerConfigError.
	envConfigMapEnvVar = "TFX_EXECUTOR_ENV_CONFIGMAP"
	envKeyEnvVar       = "TFX_EXECUTOR_ENV_KEY"
)

type executorConfigMaps struct {
	volumeConfigMap string
	volumeMountPath string
	envConfigMap    string
	envKey          string
}

func (c executorConfigMaps) hasVolume() bool {
	return c.volumeConfigMap != ""
}

func (c executorConfigMaps) hasEnv() bool {
	return c.envConfigMap != ""
}

func executorConfigMapsFromEnv() (executorConfigMaps, error) {
	c := executorConfigMaps{
		volumeConfigMap: os.Getenv(volumeConfigMapEnvVar),
		volumeMountPath: os.Getenv(volumeMountPathEnvVar),
		envConfigMap:    os.Getenv(envConfigMapEnvVar),
		envKey:          os.Getenv(envKeyEnvVar),
	}

	if err := requireTogether(volumeConfigMapEnvVar, c.volumeConfigMap, volumeMountPathEnvVar, c.volumeMountPath); err != nil {
		return executorConfigMaps{}, err
	}
	if err := requireTogether(envConfigMapEnvVar, c.envConfigMap, envKeyEnvVar, c.envKey); err != nil {
		return executorConfigMaps{}, err
	}
	return c, nil
}

func requireTogether(nameA, valueA, nameB, valueB string) error {
	if (valueA == "") != (valueB == "") {
		return fmt.Errorf("%s and %s must be set together", nameA, nameB)
	}
	return nil
}

// executorSpec returns the platform spec of a single container executor.
func (c executorConfigMaps) executorSpec() map[string]any {
	spec := map[string]any{}
	if c.hasVolume() {
		spec["configMapAsVolume"] = []any{
			map[string]any{
				"configMapName": c.volumeConfigMap,
				"mountPath":     c.volumeMountPath,
				"optional":      false,
			},
		}
	}
	if c.hasEnv() {
		spec["configMapAsEnv"] = []any{
			map[string]any{
				"configMapName": c.envConfigMap,
				"keyToEnv": []any{
					map[string]any{
						"configMapKey": c.envKey,
						"envVar":       c.envKey,
					},
				},
			},
		}
	}
	return spec
}

// addHackedTFXPlatformSpec replaces the PlatformSpec document of a TFX pipeline
// so that every container executor mounts the configured ConfigMap volume and
// exposes the configured ConfigMap key as an env var. Pipelines of other
// frameworks, or with nothing configured, are returned unchanged.
func addHackedTFXPlatformSpec(
	pipeline resource.CompiledPipeline,
	frameworkName string,
) (resource.CompiledPipeline, error) {
	if !strings.EqualFold(frameworkName, "tfx") {
		return pipeline, nil
	}

	configMaps, err := executorConfigMapsFromEnv()
	if err != nil {
		return nil, err
	}
	if !configMaps.hasVolume() && !configMaps.hasEnv() {
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
		return pipeline, nil
	}

	executors := make(map[string]any, len(executorNames))
	for _, name := range executorNames {
		executors[name] = configMaps.executorSpec()
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
