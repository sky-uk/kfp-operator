package v1alpha4

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
	Schedule          string             `json:"schedule,omitempty"`
	RuntimeParameters []apis.NamedValue  `json:"runtimeParameters,omitempty"`
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

	Spec   RunScheduleSpec `json:"spec,omitempty"`
	Status Status          `json:"status,omitempty"`
}

func (rc RunSchedule) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(rc.Spec.Pipeline.String())
	oh.WriteStringField(rc.Spec.ExperimentName)
	oh.WriteStringField(rc.Spec.Schedule)
	pipelines.WriteKVListField(oh, rc.Spec.RuntimeParameters)
	return oh.Sum()
}

func (rc RunSchedule) ComputeVersion() string {
	hash := rc.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

func (rc *RunSchedule) GetPipeline() PipelineIdentifier {
	return rc.Spec.Pipeline
}

func (rc *RunSchedule) GetStatus() Status {
	return rc.Status
}

func (rc *RunSchedule) SetStatus(status Status) {
	rc.Status = status
}

func (rc *RunSchedule) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      rc.Name,
		Namespace: rc.Namespace,
	}
}

func (rc *RunSchedule) GetKind() string {
	return "runSchedule"
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
