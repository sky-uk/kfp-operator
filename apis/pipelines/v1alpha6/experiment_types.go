package v1alpha6

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ExperimentSpec struct {
	Provider    common.NamespacedName `json:"provider" yaml:"provider"`
	Description string                `json:"description,omitempty"`
}

func (es Experiment) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(es.Spec.Description)
	return oh.Sum()
}

func (es Experiment) ComputeVersion() string {
	hash := es.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName="mlexp"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider.name"
// +kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
// +kubebuilder:storageversion
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

func (e Experiment) GetKind() string {
	return "experiment"
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
