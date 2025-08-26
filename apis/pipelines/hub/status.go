package v1beta1

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/pkg/common"
)

// +kubebuilder:object:generate=true
type ProviderAndId struct {
	Name common.NamespacedName `json:"name,omitempty"`
	Id   string                `json:"id,omitempty"`
}

// +kubebuilder:object:generate=true
type Status struct {
	Provider           ProviderAndId   `json:"provider,omitempty"`
	Version            string          `json:"version,omitempty"`
	ObservedGeneration int64           `json:"observedGeneration,omitempty"`
	Conditions         apis.Conditions `json:"conditions,omitempty"`
}
