package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProviderSpec struct {
	Image         string `json:"image" yaml:"image"`
	ExecutionMode string `json:"executionMode" yaml:"executionMode"`
	// +kubebuilder:validation:Pattern:=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	ServiceAccount  string            `json:"serviceAccount" yaml:"serviceAccount"`
	DefaultBeamArgs []apis.NamedValue `json:"defaultBeamArgs,omitempty" yaml:"defaultBeamArgs,omitempty"`
	// +kubebuilder:validation:Pattern:=`^[a-z]+:\/\/[a-z0-9\-_\/\.]+$`
	PipelineRootStorage string            `json:"pipelineRootStorage" yaml:"pipelineRootStorage"`
	Parameters          map[string]string `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="mlpr"
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.conditions[?(@.type == 'Synchronized')].reason"
//+kubebuilder:storageversion

type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderSpec   `json:"spec,omitempty"`
	Status ProviderStatus `json:"status,omitempty"`
}

type ProviderStatus struct {
	Conditions Conditions `json:"conditions,omitempty"`
}

func (ps *ProviderStatus) SetSynchronizationState(state apis.SynchronizationState, message string, observedGeneration int64) {
	condition := metav1.Condition{
		Type:               ConditionTypes.SynchronizationSucceeded,
		Message:            message,
		ObservedGeneration: observedGeneration,
		Reason:             string(state),
		LastTransitionTime: metav1.Now(),
		Status:             ConditionStatusForSynchronizationState(state),
	}
	ps.Conditions = ps.Conditions.MergeIntoConditions(condition)
}
