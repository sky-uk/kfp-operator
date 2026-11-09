package provider

import (
	"fmt"
	"sort"
	"strings"

	yamlv3 "go.yaml.in/yaml/v3"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

// HACK: hard-coded volume configuration for TFX pipelines.
//
// This is a temporary, deliberately inelegant mechanism to verify that the
// executor pods KFP spins up can be given an extra volume + volume mount (used
// to mount a ConfigMap). Replace the placeholder values below with the real
// ones once confirmed. If this works, it should be replaced by a proper,
// configurable mechanism.
//
// NOTE: the KFP driver uses the ConfigMap name verbatim as the pod volume name,
// so it must be a valid lowercase DNS-1123 label (no uppercase, no underscores).
const (
	hackConfigMapName = "foo-config"
	hackMountPath     = "/etc/placeholder"
)

// addHackedTFXPlatformSpec appends (or replaces) a KFP PlatformSpec document
// that mounts hackConfigMapName at hackMountPath into every container executor
// of the pipeline, when the framework is TFX. Other frameworks are returned
// unchanged.
//
// A KFP IR YAML is a PipelineSpec followed by an optional PlatformSpec as a
// second document. The KFP driver reads
// platforms.kubernetes.deploymentSpec.executors.<executor>.configMapAsVolume
// and adds both the pod Volume and the container VolumeMount for it.
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
		// Nothing to mount into; leave the pipeline as it is.
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

	// Keep the PipelineSpec only; any existing PlatformSpec is replaced.
	encoded, err := encodeYAMLDocuments([]*yamlv3.Node{documents[0], &platformSpec})
	if err != nil {
		return nil, err
	}
	return resource.CompiledPipeline(encoded), nil
}

// containerExecutorNames returns the sorted names of the executors in
// deploymentSpec.executors that run a user container. Importer and resolver
// executors are skipped as they don't run in an executor pod of their own.
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
