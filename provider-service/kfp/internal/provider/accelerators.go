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
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
	corev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"
)

// Accelerator says how to place an executor that requests an accelerator type.
type Accelerator struct {
	// ResourceType is the Kubernetes resource that replaces the requested
	// accelerator type as the executor's container limit, e.g. nvidia.com/gpu.
	ResourceType string `json:"resourceType"`
	// NodeSelector is added to the executor unless it already has one.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations are added to the executor's tolerations.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// AcceleratorConfig maps each accelerator type a pipeline can request to how
// its executors are placed.
type AcceleratorConfig struct {
	Accelerators map[string]Accelerator `json:"accelerators"`
}

// LoadAcceleratorConfig reads an AcceleratorConfig from a YAML file.
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

// Apply places every executor that requests an accelerator. The requested type
// is rewritten to the configured resource type, and the configured node
// selector and tolerations are added to the executor's Kubernetes platform
// spec. An executor that already has a node selector keeps it.
//
// A requested type that is not configured fails, unless the executor already
// has a node selector. A nil or empty config changes nothing.
func (c *AcceleratorConfig) Apply(pipelinePackage *PipelinePackage) error {
	if c == nil || len(c.Accelerators) == 0 {
		return nil
	}

	deploymentConfig, err := pipelinePackage.DeploymentConfig()
	if err != nil {
		return err
	}

	for _, executor := range slices.Sorted(maps.Keys(deploymentConfig.GetExecutors())) {
		accelerator := deploymentConfig.GetExecutors()[executor].GetContainer().GetResources().GetAccelerator()
		requested, ok := requestedAcceleratorType(accelerator)
		if !ok {
			continue
		}

		configured, ok := c.Accelerators[requested]
		if !ok {
			if hasNodeSelector(pipelinePackage.KubernetesExecutorConfig(executor)) {
				continue
			}
			if strings.Contains(requested, "{{$") {
				return fmt.Errorf("executor %q requests accelerator type %q, which is only resolved at run time; set a fixed accelerator type", executor, requested)
			}
			return fmt.Errorf(
				"executor %q requests accelerator type %q, which is not configured; configured types: %s",
				executor, requested, strings.Join(slices.Sorted(maps.Keys(c.Accelerators)), ", "),
			)
		}

		pipelinePackage.SetAcceleratorType(executor, configured.ResourceType)
		if err := configured.place(pipelinePackage, executor); err != nil {
			return fmt.Errorf("failed to place executor %q: %w", executor, err)
		}
	}
	return nil
}

// place adds the node selector and tolerations to the executor's Kubernetes
// platform spec.
func (a Accelerator) place(pipelinePackage *PipelinePackage, executor string) error {
	if len(a.NodeSelector) == 0 && len(a.Tolerations) == 0 {
		return nil
	}
	executorConfig := pipelinePackage.EnsureKubernetesExecutorConfig(executor)

	if len(a.NodeSelector) > 0 && !hasNodeSelector(executorConfig) {
		nodeSelector, err := structValue(&kubernetesplatform.NodeSelector{Labels: a.NodeSelector})
		if err != nil {
			return err
		}
		executorConfig.Fields["nodeSelector"] = nodeSelector
	}

	if len(a.Tolerations) == 0 {
		return nil
	}
	tolerations := &structpb.ListValue{}
	if existing, ok := executorConfig.Fields["tolerations"]; ok {
		if existing.GetListValue() == nil {
			return fmt.Errorf("tolerations is not a list")
		}
		tolerations = existing.GetListValue()
	}
	for _, t := range a.Tolerations {
		toleration, err := structValue(&kubernetesplatform.Toleration{
			Key:               t.Key,
			Operator:          string(t.Operator),
			Value:             t.Value,
			Effect:            string(t.Effect),
			TolerationSeconds: t.TolerationSeconds,
		})
		if err != nil {
			return err
		}
		if !slices.ContainsFunc(tolerations.Values, func(v *structpb.Value) bool { return proto.Equal(v, toleration) }) {
			tolerations.Values = append(tolerations.Values, toleration)
		}
	}
	executorConfig.Fields["tolerations"] = structpb.NewListValue(tolerations)
	return nil
}

// requestedAcceleratorType returns the accelerator type an executor requests.
// It reads the request the way the Kubeflow Pipelines driver does: an executor
// gets an accelerator when it has both a type and a count, and the count is
// not zero.
func requestedAcceleratorType(
	accelerator *pipelinespec.PipelineDeploymentConfig_PipelineContainerSpec_ResourceSpec_AcceleratorConfig,
) (string, bool) {
	acceleratorType := accelerator.GetResourceType()
	if acceleratorType == "" {
		acceleratorType = accelerator.GetType()
	}
	count := accelerator.GetResourceCount()
	if count == "" && accelerator.GetCount() > 0 {
		count = strconv.FormatInt(accelerator.GetCount(), 10)
	}
	if acceleratorType == "" || count == "" {
		return "", false
	}
	if quantity, err := k8sresource.ParseQuantity(count); err == nil && quantity.IsZero() {
		return "", false
	}
	return acceleratorType, true
}

func hasNodeSelector(executorConfig *structpb.Struct) bool {
	_, camelCase := executorConfig.GetFields()["nodeSelector"]
	_, snakeCase := executorConfig.GetFields()["node_selector"]
	return camelCase || snakeCase
}

func structValue(message proto.Message) (*structpb.Value, error) {
	asJSON, err := protojson.Marshal(message)
	if err != nil {
		return nil, err
	}
	value := &structpb.Value{}
	if err := protojson.Unmarshal(asJSON, value); err != nil {
		return nil, err
	}
	return value, nil
}
