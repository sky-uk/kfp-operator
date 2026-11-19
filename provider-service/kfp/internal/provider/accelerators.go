package provider

import (
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/kubeflow/pipelines/api/v2alpha1/go/pipelinespec"
	"github.com/kubeflow/pipelines/kubernetes_platform/go/kubernetesplatform"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"google.golang.org/protobuf/types/known/structpb"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

// placeAccelerators sets the configured placement on each executor that requests an accelerator.
func placeAccelerators(accelerators config.Accelerators, pipelinePackage *PipelinePackage) error {
	if len(accelerators) == 0 {
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

		configured, isConfigured := accelerators[requested]
		if !isConfigured {
			settings, err := pipelinePackage.KubernetesExecutorConfig(executor)
			if err != nil {
				return err
			}
			if settings.GetNodeSelector() != nil {
				continue
			}
			return unconfiguredTypeError(accelerators, executor, requested)
		}

		setAcceleratorType(request, configured.ResourceType)
		if err := pipelinePackage.MergeKubernetesExecutorConfig(executor, placement(configured)); err != nil {
			return fmt.Errorf("failed to place executor %q: %w", executor, err)
		}
	}
	return nil
}

func unconfiguredTypeError(accelerators config.Accelerators, executor string, requested string) error {
	if strings.Contains(requested, "{{$") {
		return fmt.Errorf(
			"executor %q requests accelerator type %q, which is only resolved at run time; set a fixed accelerator type",
			executor, requested,
		)
	}
	return fmt.Errorf(
		"executor %q requests accelerator type %q, which is not configured; configured types: %s",
		executor, requested, strings.Join(slices.Sorted(maps.Keys(accelerators)), ", "),
	)
}

func placement(accelerator config.Accelerator) *kubernetesplatform.KubernetesExecutorConfig {
	settings := &kubernetesplatform.KubernetesExecutorConfig{}
	if len(accelerator.NodeSelector) > 0 {
		settings.NodeSelector = &kubernetesplatform.NodeSelector{Labels: accelerator.NodeSelector}
	}
	for _, t := range accelerator.Tolerations {
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
