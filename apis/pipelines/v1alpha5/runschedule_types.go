package v1alpha5

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (rs RunSchedule) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(rs.Spec.Pipeline.String())
	oh.WriteStringField(rs.Spec.ExperimentName)
	pipelines.WriteKVListField(oh, rs.Spec.RuntimeParameters)
	pipelines.WriteKVListField(oh, rs.Spec.Artifacts)
	oh.WriteStringField(rs.Spec.Schedule)
	return oh.Sum()
}

func (rs RunSchedule) ComputeVersion() string {
	hash := rs.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="mlrs"
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ProviderId",type="string",JSONPath=".status.providerId"
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"

type RunSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   hub.RunScheduleSpec `json:"spec,omitempty"`
	Status Status              `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type RunScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RunSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RunSchedule{}, &RunScheduleList{})
}
