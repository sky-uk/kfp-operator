package config

import (
	"encoding/json"
	"fmt"

	"github.com/kubeflow/pipelines/kubernetes_platform/go/kubernetesplatform"
	"google.golang.org/protobuf/encoding/protojson"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

const acceleratorsField = "accelerators"

// Platform holds the settings the provider adds to pipeline executors.
type Platform struct {
	Accelerators Accelerators
	Kubernetes   *kubernetesplatform.KubernetesExecutorConfig
}

// Accelerators maps accelerator types to where their executors run.
type Accelerators map[string]Accelerator

type Accelerator struct {
	ResourceType string              `json:"resourceType"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
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
	for name, accelerator := range decoded.Accelerators {
		if accelerator.ResourceType == "" {
			return fmt.Errorf("invalid platform: accelerator %q has no resourceType", name)
		}
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
