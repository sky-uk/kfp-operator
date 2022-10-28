package v1alpha4

import (
	. "github.com/sky-uk/kfp-operator/apis"
)

// +kubebuilder:object:generate=true
type Status struct {
	ProviderId           string               `json:"providerId,omitempty"`
	SynchronizationState SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string               `json:"version,omitempty"`
	ObservedGeneration   int64                `json:"observedGeneration,omitempty"`
}
