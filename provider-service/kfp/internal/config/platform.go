package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kubeflow/pipelines/kubernetes_platform/go/kubernetesplatform"
	"github.com/samber/lo"
	"google.golang.org/protobuf/encoding/protojson"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/yaml"
)

const (
	acceleratorsField     = "accelerators"
	acceleratorsFromField = "acceleratorsFrom"
)

var configMapsResource = schema.GroupVersionResource{Version: "v1", Resource: "configmaps"}

// Platform holds the settings the provider adds to pipeline executors.
type Platform struct {
	Accelerators     Accelerators
	AcceleratorsFrom *AcceleratorsSource
	Kubernetes       *kubernetesplatform.KubernetesExecutorConfig
}

// Accelerators maps accelerator types to where their executors run.
type Accelerators map[string]Accelerator

type Accelerator struct {
	ResourceType string              `json:"resourceType"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

// AcceleratorsSource is a key of a ConfigMap in the provider's namespace that holds accelerators as YAML.
type AcceleratorsSource struct {
	ConfigMapKeyRef *corev1.ConfigMapKeySelector `json:"configMapKeyRef"`
}

// MarshalText seeds the parameter as one value, which viper needs to read it from the environment.
func (p Platform) MarshalText() ([]byte, error) {
	fields := map[string]any{}
	if p.Kubernetes != nil {
		settings, err := protojson.Marshal(p.Kubernetes)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(settings, &fields); err != nil {
			return nil, err
		}
	}
	if len(p.Accelerators) > 0 {
		fields[acceleratorsField] = p.Accelerators
	}
	if p.AcceleratorsFrom != nil {
		fields[acceleratorsFromField] = p.AcceleratorsFrom
	}
	return json.Marshal(fields)
}

// UnmarshalText decodes the JSON itself, as viper lowercases keys and splits them on dots.
func (p *Platform) UnmarshalText(text []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(text, &fields); err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}
	if fields == nil {
		fields = map[string]json.RawMessage{}
	}

	decoded := Platform{Kubernetes: &kubernetesplatform.KubernetesExecutorConfig{}}
	if accelerators, isSet := fields[acceleratorsField]; isSet {
		if err := yaml.UnmarshalStrict(accelerators, &decoded.Accelerators); err != nil {
			return fmt.Errorf("invalid platform: %w", err)
		}
		delete(fields, acceleratorsField)
	}
	if err := validateAccelerators(decoded.Accelerators); err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}
	if source, isSet := fields[acceleratorsFromField]; isSet {
		if err := yaml.UnmarshalStrict(source, &decoded.AcceleratorsFrom); err != nil {
			return fmt.Errorf("invalid platform: %w", err)
		}
		if decoded.AcceleratorsFrom == nil || decoded.AcceleratorsFrom.ConfigMapKeyRef == nil ||
			decoded.AcceleratorsFrom.ConfigMapKeyRef.Name == "" || decoded.AcceleratorsFrom.ConfigMapKeyRef.Key == "" {
			return fmt.Errorf("invalid platform: %s needs a configMapKeyRef with a name and a key", acceleratorsFromField)
		}
		delete(fields, acceleratorsFromField)
	}

	settings, err := json.Marshal(fields)
	if err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}
	if err := protojson.Unmarshal(settings, decoded.Kubernetes); err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}

	*p = decoded
	return nil
}

// ResolveAccelerators adds the accelerators from AcceleratorsFrom. Inline accelerators take precedence by type.
func (p *Platform) ResolveAccelerators(ctx context.Context, client dynamic.Interface, namespace string) error {
	if p.AcceleratorsFrom == nil {
		return nil
	}
	ref := p.AcceleratorsFrom.ConfigMapKeyRef
	optional := ptr.Deref(ref.Optional, false)

	configMap, err := client.Resource(configMapsResource).Namespace(namespace).Get(ctx, ref.Name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) && optional {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to get accelerators ConfigMap %q: %w", ref.Name, err)
	}
	data, _, err := unstructured.NestedStringMap(configMap.Object, "data")
	if err != nil {
		return fmt.Errorf("invalid accelerators ConfigMap %q: %w", ref.Name, err)
	}
	value, hasKey := data[ref.Key]
	if !hasKey {
		if optional {
			return nil
		}
		return fmt.Errorf("accelerators ConfigMap %q has no key %q", ref.Name, ref.Key)
	}

	var accelerators Accelerators
	if err := yaml.UnmarshalStrict([]byte(value), &accelerators); err != nil {
		return fmt.Errorf("invalid accelerators in ConfigMap %q: %w", ref.Name, err)
	}
	if err := validateAccelerators(accelerators); err != nil {
		return fmt.Errorf("invalid accelerators in ConfigMap %q: %w", ref.Name, err)
	}
	p.Accelerators = lo.Assign(accelerators, p.Accelerators)
	return nil
}

func validateAccelerators(accelerators Accelerators) error {
	if name, hasNoResourceType := lo.FindKeyBy(accelerators, func(_ string, accelerator Accelerator) bool {
		return accelerator.ResourceType == ""
	}); hasNoResourceType {
		return fmt.Errorf("accelerator %q has no resourceType", name)
	}
	return nil
}
