package v1

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/controllers/objecthasher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PipelineSpec struct {
	Image         string            `json:"image" yaml:"image"`
	TfxComponents string            `json:"tfxComponents" yaml:"tfxComponents"`
	Env           map[string]string `json:"env,omitempty" yaml:"env"`
	BeamArgs      map[string]string `json:"beamArgs,omitempty" yaml:"beamArgs"`
}

func (ps PipelineSpec) ComputeVersion() string {
	oh := objecthasher.New()
	oh.WriteStringField(ps.Image)
	oh.WriteStringField(ps.TfxComponents)
	oh.WriteMapField(ps.Env)
	oh.WriteMapField(ps.BeamArgs)
	specHash := oh.Sum()

	return fmt.Sprintf("%x", specHash)
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
	Version              string               `json:"version,omitempty"`
	SynchronizationState SynchronizationState `json:"synchronizationState,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="PipelineId",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"

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
