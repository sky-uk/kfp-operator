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

type Accelerator struct {
	ResourceType string              `json:"resourceType"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

// AcceleratorConfig maps accelerator types to where their executors run.
type AcceleratorConfig struct {
	Accelerators map[string]Accelerator `json:"accelerators"`
}

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

// Apply sets the configured placement on each executor that requests an accelerator.
func (c *AcceleratorConfig) Apply(pipelinePackage *PipelinePackage) error {
	if c == nil || len(c.Accelerators) == 0 {
		return nil
	}

	executors := nestedStruct(pipelinePackage.PipelineSpec.GetDeploymentSpec(), "executors").GetFields()
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

func (c *AcceleratorConfig) unconfiguredTypeError(executor string, requested string) error {
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

// requestedAcceleratorType follows the KFP driver, which needs a type and a non-zero count.
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
	if quantity, err := k8sresource.ParseQuantity(count); err == nil && quantity.IsZero() {
		return "", false, nil
	}
	return acceleratorType, true, nil
}

func setAcceleratorType(request *structpb.Struct, acceleratorType string) {
	for _, name := range []string{"resourceType", "resource_type", "type"} {
		if request.Fields[name].GetStringValue() != "" {
			request.Fields[name] = structpb.NewStringValue(acceleratorType)
		}
	}
}
