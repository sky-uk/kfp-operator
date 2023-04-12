package v1alpha3

import (
	. "github.com/sky-uk/kfp-operator/apis"
)

// +kubebuilder:object:generate=true
type Status struct {
	KfpId                string               `json:"kfpId,omitempty"`
	SynchronizationState SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string               `json:"version,omitempty"`
	ObservedGeneration   int64                `json:"observedGeneration,omitempty"`
}
