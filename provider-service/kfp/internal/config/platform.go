package config

import (
	"encoding/json"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

// Platform holds the settings the provider adds to pipeline executors.
type Platform struct {
	Accelerators Accelerators `json:"accelerators,omitempty"`
}

// Accelerators maps accelerator types to where their executors run.
type Accelerators map[string]Accelerator

type Accelerator struct {
	ResourceType string              `json:"resourceType"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

type platform Platform

// MarshalText seeds the parameter as one value, which viper needs to read it from the environment.
func (p Platform) MarshalText() ([]byte, error) {
	return json.Marshal(platform(p))
}

// UnmarshalText decodes the JSON itself, as viper lowercases keys and splits them on dots.
func (p *Platform) UnmarshalText(text []byte) error {
	var decoded platform
	if err := yaml.UnmarshalStrict(text, &decoded); err != nil {
		return fmt.Errorf("invalid platform: %w", err)
	}
	for name, accelerator := range decoded.Accelerators {
		if accelerator.ResourceType == "" {
			return fmt.Errorf("invalid platform: accelerator %q has no resourceType", name)
		}
	}
	*p = Platform(decoded)
	return nil
}
