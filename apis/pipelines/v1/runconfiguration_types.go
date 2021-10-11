package v1

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/controllers/objecthasher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type RunConfigurationSpec struct {
	PipelineName string `json:"pipelineName,omitempty"`
	Schedule string `json:"schedule,omitempty"`
	RuntimeParameters map[string]string `json:"runtimeParameters,omitempty"`
}

func (rcs RunConfigurationSpec) ComputeHash() []byte {
	oh := objecthasher.New()
	oh.WriteStringField(rcs.PipelineName)
	oh.WriteStringField(rcs.Schedule)
	oh.WriteMapField(rcs.RuntimeParameters)
	return oh.Sum()
}

func (rcs RunConfigurationSpec) ComputeVersion() string {
	hash := rcs.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

type RunConfigurationStatus struct {
	Id                   string               `json:"id,omitempty"`
	Version              string               `json:"version,omitempty"`
	SynchronizationState SynchronizationState `json:"synchronizationState,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Status",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="PipelineId",type="string",JSONPath=".status.id"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"

type RunConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunConfigurationSpec   `json:"spec,omitempty"`
	Status RunConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

type RunConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RunConfiguration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RunConfiguration{}, &RunConfigurationList{})
}
