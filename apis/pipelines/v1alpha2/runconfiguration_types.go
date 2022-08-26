package v1alpha2

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RunConfigurationSpec struct {
	Pipeline          PipelineIdentifier `json:"pipeline,omitempty"`
	ExperimentName    string             `json:"experimentName,omitempty"`
	Schedule          string             `json:"schedule,omitempty"`
	RuntimeParameters map[string]string  `json:"runtimeParameters,omitempty"`
}

func (rc RunConfiguration) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(rc.Spec.Pipeline.String())
	oh.WriteStringField(rc.Spec.ExperimentName)
	oh.WriteStringField(rc.Spec.Schedule)
	oh.WriteMapField(rc.Spec.RuntimeParameters)
	oh.WriteStringField(rc.Status.ObservedPipelineVersion)
	return oh.Sum()
}

func (rc RunConfiguration) ComputeVersion() string {
	hash := rc.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

type RunConfigurationStatus struct {
	apis.Status             `json:",inline"`
	ObservedPipelineVersion string `json:"observedPipelineVersion,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="mlrc"
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="KfpId",type="string",JSONPath=".status.kfpId"
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"

type RunConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunConfigurationSpec   `json:"spec,omitempty"`
	Status RunConfigurationStatus `json:"status,omitempty"`
}

func (rc *RunConfiguration) GetStatus() apis.Status {
	return rc.Status.Status
}

func (rc *RunConfiguration) SetStatus(status apis.Status) {
	rc.Status.Status = status
}

func (rc RunConfiguration) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      rc.Name,
		Namespace: rc.Namespace,
	}
}

func (rc RunConfiguration) GetKind() string {
	return "runconfiguration"
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
