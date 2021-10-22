package v1

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/controllers/objecthasher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RunConfigurationSpec struct {
	PipelineName      string            `json:"pipelineName,omitempty"`
	Schedule          string            `json:"schedule,omitempty"`
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="KfpId",type="string",JSONPath=".status.kfpId"
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"

type RunConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunConfigurationSpec `json:"spec,omitempty"`
	Status Status               `json:"status,omitempty"`
}

func (r RunConfiguration) NamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
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
