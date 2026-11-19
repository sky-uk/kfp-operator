package provider

import (
	"errors"
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

// Accelerator describes how to place an executor that requests an accelerator
// type.
type Accelerator struct {
	// ResourceType is the Kubernetes resource that replaces the requested
	// accelerator type as the container's limit, e.g. nvidia.com/gpu.
	ResourceType string `json:"resourceType"`
	// NodeSelector is added to the executor unless it already has one.
	NodeSelector map[string]string `json:"nodeSelector,omitempty"`
	// Tolerations are added to the executor's tolerations.
	Tolerations []corev1.Toleration `json:"tolerations,omitempty"`
}

// AcceleratorConfig maps each accelerator type a pipeline can request to how
// executors requesting it are placed.
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

// Apply places every executor that requests an accelerator. The requested type
// is replaced with the configured resource type, and the configured node
// selector and tolerations are added to the executor's Kubernetes settings. An
// executor that already has a node selector keeps it.
//
// A requested type that is not configured fails, unless the executor already
// has a node selector. A nil or empty config changes nothing.
func (c *AcceleratorConfig) Apply(pipelinePackage *PipelinePackage) error {
	if c == nil || len(c.Accelerators) == 0 {
		return nil
	}

	deploymentConfig, err := pipelinePackage.deploymentConfig()
	if err != nil {
		return err
	}

	// Sorted so that a pipeline with several unplaceable executors always
	// reports the same one.
	for _, executor := range slices.Sorted(maps.Keys(deploymentConfig.GetExecutors())) {
		request := deploymentConfig.GetExecutors()[executor].GetContainer().GetResources().GetAccelerator()
		requested, isRequested := requestedAcceleratorType(request)
		if !isRequested {
			continue
		}

		configured, isConfigured := c.Accelerators[requested]
		if !isConfigured {
			if hasNodeSelector(pipelinePackage.kubernetesExecutorConfig(executor)) {
				continue
			}
			// KFP SDK writes a type taken from a pipeline parameter as a
			// placeholder, which the driver only resolves when the task runs.
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

		pipelinePackage.setAcceleratorType(executor, configured.ResourceType)
		if err := configured.place(pipelinePackage, executor); err != nil {
			return fmt.Errorf("failed to place executor %q: %w", executor, err)
		}
	}
	return nil
}

// place adds the node selector and tolerations to the executor's Kubernetes
// settings.
func (a Accelerator) place(pipelinePackage *PipelinePackage, executor string) error {
	if len(a.NodeSelector) == 0 && len(a.Tolerations) == 0 {
		return nil
	}
	executorConfig := pipelinePackage.ensureKubernetesExecutorConfig(executor)

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
	if existing, hasTolerations := executorConfig.Fields["tolerations"]; hasTolerations {
		if existing.GetListValue() == nil {
			return errors.New("tolerations is not a list")
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
// It follows the Kubeflow Pipelines driver, which only sets an accelerator
// limit when both a type and a non-zero count are set, and prefers
// resourceType and resourceCount over the deprecated type and count.
func requestedAcceleratorType(
	request *pipelinespec.PipelineDeploymentConfig_PipelineContainerSpec_ResourceSpec_AcceleratorConfig,
) (string, bool) {
	acceleratorType := request.GetResourceType()
	if acceleratorType == "" {
		acceleratorType = request.GetType()
	}
	count := request.GetResourceCount()
	if count == "" && request.GetCount() > 0 {
		count = strconv.FormatInt(request.GetCount(), 10)
	}
	if acceleratorType == "" || count == "" {
		return "", false
	}
	// A count that is not a quantity, such as a run-time placeholder, may still
	// resolve to a non-zero count, so it counts as a request.
	if quantity, err := k8sresource.ParseQuantity(count); err == nil && quantity.IsZero() {
		return "", false
	}
	return acceleratorType, true
}

// hasNodeSelector reports whether the executor's Kubernetes settings set a node
// selector, under either the JSON or the proto field name that protojson reads.
func hasNodeSelector(executorConfig *structpb.Struct) bool {
	_, hasJSONName := executorConfig.GetFields()["nodeSelector"]
	_, hasProtoName := executorConfig.GetFields()["node_selector"]
	return hasJSONName || hasProtoName
}

// structValue converts a message to the untyped value Kubernetes settings are
// held as.
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
