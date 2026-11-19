package config

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

type Accelerator struct {
	ResourceType string              `json:"resourceType"`
	NodeSelector map[string]string   `json:"nodeSelector,omitempty"`
	Tolerations  []corev1.Toleration `json:"tolerations,omitempty"`
}

// Accelerators maps accelerator types to where their executors run.
type Accelerators map[string]Accelerator

// UnmarshalText decodes the JSON itself, as viper lowercases keys and splits them on dots.
func (a *Accelerators) UnmarshalText(text []byte) error {
	accelerators := map[string]Accelerator{}
	if err := yaml.UnmarshalStrict(text, &accelerators); err != nil {
		return fmt.Errorf("invalid accelerators: %w", err)
	}
	for name, accelerator := range accelerators {
		if accelerator.ResourceType == "" {
			return fmt.Errorf("invalid accelerators: %q has no resourceType", name)
		}
	}
	*a = accelerators
	return nil
}
