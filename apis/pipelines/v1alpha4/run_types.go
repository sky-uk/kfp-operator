package v1alpha4

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RunSpec struct {
	Pipeline          PipelineIdentifier `json:"pipeline,omitempty"`
	ExperimentName    string             `json:"experimentName,omitempty"`
	RuntimeParameters []apis.NamedValue  `json:"runtimeParameters,omitempty"`
}

func (r Run) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(r.Spec.Pipeline.String())
	oh.WriteStringField(r.Spec.ExperimentName)
	oh.WriteNamedValueListField(r.Spec.RuntimeParameters)
	return oh.Sum()
}

func (r Run) ComputeVersion() string {
	hash := r.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

type CompletionState string

var CompletionStates = struct {
	Succeeded CompletionState
	Failed CompletionState
}{
	Succeeded: "Succeeded",
	Failed: "Failed",
}

type RunStatus struct {
	CompletionState CompletionState `json:"completionState,omitempty"`
	Status `json:",inline"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="mlr"
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ProviderId",type="string",JSONPath=".status.providerId"
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
//+kubebuilder:printcolumn:name="CompletionState",type="string",JSONPath=".status.completionState"
//+kubebuilder:storageversion

type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunSpec `json:"spec,omitempty"`
	Status RunStatus  `json:"status,omitempty"`
}

func (r *Run) GetStatus() Status {
	return r.Status.Status
}

func (r *Run) SetStatus(status Status) {
	r.Status.Status = status
}

func (r Run) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}

func (r Run) GetKind() string {
	return "run"
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
