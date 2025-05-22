package v1alpha5

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RunScheduleSpec struct {
	Pipeline          PipelineIdentifier `json:"pipeline,omitempty"`
	ExperimentName    string             `json:"experimentName,omitempty"`
	RuntimeParameters []apis.NamedValue  `json:"runtimeParameters,omitempty"`
	// Needed for conversion only
	// +kubebuilder:validation:-
	// +optional
	Parameters []apis.NamedValue `json:"parameters,omitempty"`
	Artifacts  []OutputArtifact  `json:"artifacts,omitempty"`
	Schedule   string            `json:"schedule,omitempty"`
}

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
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.conditions[?(@.type==\"Synchronized\")].reason"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"

type RunSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunScheduleSpec `json:"spec,omitempty"`
	Status Status          `json:"status,omitempty"`
}

func (rs *RunSchedule) GetProvider() string {
	return rs.Status.ProviderId.Provider
}

func (rs *RunSchedule) GetPipeline() PipelineIdentifier {
	return rs.Spec.Pipeline
}

func (rs *RunSchedule) GetStatus() Status {
	return rs.Status
}

func (rs *RunSchedule) SetStatus(status Status) {
	rs.Status = status
}

func (rs *RunSchedule) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      rs.Name,
		Namespace: rs.Namespace,
	}
}

func (rs *RunSchedule) GetKind() string {
	return "runschedule"
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
