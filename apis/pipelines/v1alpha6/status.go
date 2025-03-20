package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis"
)

// +kubebuilder:object:generate=true
type ProviderAndId struct {
	Name string `json:"name,omitempty"`
	Id   string `json:"id,omitempty"`
}

// +kubebuilder:object:generate=true
type Status struct {
	Provider             ProviderAndId             `json:"provider,omitempty"`
	SynchronizationState apis.SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string                    `json:"version,omitempty"`
	ObservedGeneration   int64                     `json:"observedGeneration,omitempty"`
	Conditions           apis.Conditions           `json:"conditions,omitempty"`
}
