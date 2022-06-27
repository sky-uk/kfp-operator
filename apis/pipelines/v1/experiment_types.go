package v1

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/controllers/objecthasher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ExperimentSpec struct {
	Description string `json:"description,omitempty"`
}

func (es ExperimentSpec) ComputeHash() []byte {
	oh := objecthasher.New()
	oh.WriteStringField(es.Description)
	return oh.Sum()
}

func (es ExperimentSpec) ComputeVersion() string {
	hash := es.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName="mlexp"
//+kubebuilder:printcolumn:name="KfpId",type="string",JSONPath=".status.kfpId"
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"

type Experiment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExperimentSpec `json:"spec,omitempty"`
	Status Status         `json:"status,omitempty"`
}

func (e *Experiment) GetStatus() Status {
	return e.Status
}

func (e *Experiment) SetStatus(status Status) {
	e.Status = status
}

func (e Experiment) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      e.Name,
		Namespace: e.Namespace,
	}
}

//+kubebuilder:object:root=true

type ExperimentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Experiment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Experiment{}, &ExperimentList{})
}
