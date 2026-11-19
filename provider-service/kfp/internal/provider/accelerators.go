package provider

import (
	"fmt"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/kubeflow/pipelines/api/v2alpha1/go/pipelinespec"
	"github.com/kubeflow/pipelines/kubernetes_platform/go/kubernetesplatform"
	"google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"
)

// Accelerator describes how to place an executor that requests an accelerator
// type.
type Accelerator struct {
	// ResourceType is the Kubernetes resource that replaces the requested
	// accelerator type as the container's limit, e.g. nvidia.com/gpu.
	ResourceType string `json:"resourceType"`
	// NodeSelector goes on the executor unless the executor sets its own.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations go on the executor alongside any tolerations it sets.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// AcceleratorConfig maps each accelerator type a pipeline can request to the
// placement of the executors that request it.
type AcceleratorConfig struct {
	Accelerators map[string]Accelerator `json:"accelerators"`
}

// LoadAcceleratorConfig reads an AcceleratorConfig from a YAML file, rejecting
// unknown fields so that a misspelt field fails at start-up.
func LoadAcceleratorConfig(path string) (*AcceleratorConfig, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read accelerator config: %w", err)
	}

	acceleratorConfig := &AcceleratorConfig{}
	if err := yaml.UnmarshalStrict(contents, acceleratorConfig); err != nil {
		return nil, fmt.Errorf("invalid accelerator config %s: %w", path, err)
	}
	for name, accelerator := range acceleratorConfig.Accelerators {
		if accelerator.ResourceType == "" {
			return nil, fmt.Errorf("invalid accelerator config %s: accelerator %q has no resourceType", path, name)
		}
	}
	return acceleratorConfig, nil
}

// Apply places each executor that requests an accelerator. It replaces the
// requested type with the configured resource type and merges the configured
// node selector and tolerations into the executor's Kubernetes settings.
//
// Apply fails on a requested type the config lacks, unless the executor sets
// its own node selector. A nil or empty config changes nothing.
func (c *AcceleratorConfig) Apply(pipelinePackage *PipelinePackage) error {
	if c == nil || len(c.Accelerators) == 0 {
		return nil
	}

	executors := nestedStruct(pipelinePackage.PipelineSpec.GetDeploymentSpec(), "executors").GetFields()
	// Visit executors in sorted order so that a pipeline with several
	// unplaceable executors reports the same one on each upload.
	for _, executor := range slices.Sorted(maps.Keys(executors)) {
		request := nestedStruct(executors[executor].GetStructValue(), "container", "resources", "accelerator")
		requested, isRequested, err := requestedAcceleratorType(request)
		if err != nil {
			return fmt.Errorf("executor %q has an invalid accelerator request: %w", executor, err)
		}
		if !isRequested {
			continue
		}

		configured, isConfigured := c.Accelerators[requested]
		if !isConfigured {
			settings, err := pipelinePackage.KubernetesExecutorConfig(executor)
			if err != nil {
				return err
			}
			if settings.GetNodeSelector() != nil {
				continue
			}
			return c.unconfiguredTypeError(executor, requested)
		}

		setAcceleratorType(request, configured.ResourceType)
		if err := pipelinePackage.MergeKubernetesExecutorConfig(executor, configured.placement()); err != nil {
			return fmt.Errorf("failed to place executor %q: %w", executor, err)
		}
	}
	return nil
}

// unconfiguredTypeError reports an executor that requests an accelerator type
// with no configured placement.
func (c *AcceleratorConfig) unconfiguredTypeError(executor string, requested string) error {
	// KFP SDK writes a type taken from a pipeline parameter as a placeholder,
	// which the driver resolves when the task runs, after the upload.
	if strings.Contains(requested, "{{$") {
		return fmt.Errorf(
			"executor %q requests accelerator type %q, which is only resolved at run time; set a fixed accelerator type",
			executor, requested,
		)
	}
	return fmt.Errorf(
		"executor %q requests accelerator type %q, which is not configured; configured types: %s",
		executor, requested, strings.Join(slices.Sorted(maps.Keys(c.Accelerators)), ", "),
	)
}

// placement returns the Kubernetes settings that place an executor on the
// accelerator's nodes.
func (a Accelerator) placement() *kubernetesplatform.KubernetesExecutorConfig {
	settings := &kubernetesplatform.KubernetesExecutorConfig{}
	if len(a.NodeSelector) > 0 {
		settings.NodeSelector = &kubernetesplatform.NodeSelector{Labels: a.NodeSelector}
	}
	for _, t := range a.Tolerations {
		settings.Tolerations = append(settings.Tolerations, &kubernetesplatform.Toleration{
			Key:               t.Key,
			Operator:          string(t.Operator),
			Value:             t.Value,
			Effect:            string(t.Effect),
			TolerationSeconds: t.TolerationSeconds,
		})
	}
	return settings
}

// requestedAcceleratorType returns the accelerator type an executor's
// accelerator request asks for. It follows the Kubeflow Pipelines driver, which
// skips the accelerator limit unless the request has both a type and a non-zero
// count, and prefers resourceType and resourceCount over the deprecated type and
// count.
func requestedAcceleratorType(request *structpb.Struct) (string, bool, error) {
	if request == nil {
		return "", false, nil
	}
	accelerator := &pipelinespec.PipelineDeploymentConfig_PipelineContainerSpec_ResourceSpec_AcceleratorConfig{}
	if err := decodeStruct(request, accelerator); err != nil {
		return "", false, err
	}

	acceleratorType := accelerator.GetResourceType()
	if acceleratorType == "" {
		acceleratorType = accelerator.GetType()
	}
	count := accelerator.GetResourceCount()
	if count == "" && accelerator.GetCount() > 0 {
		count = strconv.FormatInt(accelerator.GetCount(), 10)
	}
	if acceleratorType == "" || count == "" {
		return "", false, nil
	}
	// A count that is not a quantity, such as a run-time placeholder, may
	// resolve to a non-zero count, so it counts as a request.
	if quantity, err := k8sresource.ParseQuantity(count); err == nil && quantity.IsZero() {
		return "", false, nil
	}
	return acceleratorType, true, nil
}

// setAcceleratorType replaces the type in each field of an accelerator request
// that holds one.
func setAcceleratorType(request *structpb.Struct, acceleratorType string) {
	// KFP SDK writes the deprecated type alongside resourceType, and protojson
	// reads either the JSON or the proto name of a field.
	for _, name := range []string{"resourceType", "resource_type", "type"} {
		if request.Fields[name].GetStringValue() != "" {
			request.Fields[name] = structpb.NewStringValue(acceleratorType)
		}
	}
}
