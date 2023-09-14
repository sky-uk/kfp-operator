package v1alpha5

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r Run) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	r.Spec.WriteRunSpec(oh)
	oh.WriteStringField(r.Status.ObservedPipelineVersion)
	return oh.Sum()
}

func (r Run) ComputeVersion() string {
	hash := r.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

type CompletionState string

var CompletionStates = struct {
	Succeeded CompletionState
	Failed    CompletionState
}{
	Succeeded: "Succeeded",
	Failed:    "Failed",
}

type Dependencies struct {
	RunConfigurations map[string]hub.RunReference `json:"runConfigurations,omitempty"`
}

type RunStatus struct {
	Status                  `json:",inline"`
	ObservedPipelineVersion string          `json:"observedPipelineVersion,omitempty"`
	Dependencies            Dependencies    `json:"dependencies,omitempty"`
	CompletionState         CompletionState `json:"completionState,omitempty"`
	MarkedCompletedAt       *metav1.Time    `json:"markedCompletedAt,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="mlr"
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ProviderId",type="string",JSONPath=".status.providerId"
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
//+kubebuilder:printcolumn:name="CompletionState",type="string",JSONPath=".status.completionState"

type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   hub.RunSpec `json:"spec,omitempty"`
	Status RunStatus   `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type RunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Run `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Run{}, &RunList{})
}
