package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PipelineSpec struct {
}

type SynchronizationState string

const (
	Unknown   SynchronizationState = ""
	Creating  SynchronizationState = "Creating"
	Succeeded SynchronizationState = "Succeeded"
	Updating  SynchronizationState = "Updating"
	Deleting  SynchronizationState = "Deleting"
	Deleted   SynchronizationState = "Deleted"
	Failed    SynchronizationState = "Failed"
)

type PipelineStatus struct {
	Id                   string               `json:"id,omitempty"`
	SynchronizationState SynchronizationState `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
