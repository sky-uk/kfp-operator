package v1beta1

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RunScheduleSpec struct {
	Provider       common.NamespacedName `json:"provider" yaml:"provider"`
	Pipeline       PipelineIdentifier    `json:"pipeline,omitempty"`
	ExperimentName string                `json:"experimentName,omitempty"`
	Parameters     []apis.NamedValue     `json:"parameters,omitempty"`
	// Deprecated: Needed for conversion only
	// +kubebuilder:validation:-
	// +optional
	RuntimeParameters []apis.NamedValue `json:"runtimeParameters,omitempty"`
	Artifacts         []OutputArtifact  `json:"artifacts,omitempty"`
	Schedule          Schedule          `json:"schedule,omitempty"`
}

type Schedule struct {
	CronExpression string       `json:"cronExpression,omitempty"`
	StartTime      *metav1.Time `json:"startTime,omitempty"`
	EndTime        *metav1.Time `json:"endTime,omitempty"`
}

func (s Schedule) Empty() bool {
	return s.CronExpression == "" && s.StartTime == nil && s.EndTime == nil
}

func (rs RunSchedule) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(rs.Spec.Pipeline.String())
	oh.WriteStringField(rs.Spec.ExperimentName)
	pipelines.WriteKVListField(oh, rs.Spec.Parameters)
	pipelines.WriteKVListField(oh, rs.Spec.Artifacts)
	oh.WriteStringField(rs.Spec.Schedule.CronExpression)
	if rs.Spec.Schedule.StartTime != nil {
		oh.WriteStringField(rs.Spec.Schedule.StartTime.String())
	}
	if rs.Spec.Schedule.EndTime != nil {
		oh.WriteStringField(rs.Spec.Schedule.EndTime.String())
	}
	return oh.Sum()
}

func (rs RunSchedule) ComputeVersion() string {
	hash := rs.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="mlrs"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider.name"
// +kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.conditions[?(@.type==\"Synchronized\")].reason"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
// +kubebuilder:storageversion
type RunSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunScheduleSpec `json:"spec,omitempty"`
	Status Status          `json:"status,omitempty"`
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
