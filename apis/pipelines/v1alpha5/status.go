package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ConditionTypes = struct {
	SynchronizationSucceeded string
}{
	SynchronizationSucceeded: "Synchronized",
}

func ConditionStatusForSynchronizationState(state apis.SynchronizationState) metav1.ConditionStatus {
	switch state {
	case apis.Succeeded, apis.Deleted:
		return metav1.ConditionTrue
	case apis.Failed:
		return metav1.ConditionFalse
	default:
		return metav1.ConditionUnknown
	}
}

// +kubebuilder:object:generate=true
type Status struct {
	ProviderId           hub.ProviderAndId         `json:"providerId,omitempty"`
	SynchronizationState apis.SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string                    `json:"version,omitempty"`
	ObservedGeneration   int64                     `json:"observedGeneration,omitempty"`
	Conditions           hub.Conditions            `json:"conditions,omitempty"`
}
