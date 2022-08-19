package v1alpha2

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

//```
//- name: foo
//  value: bar1
//  values:
//  - bar2
//===
//- name: foo
//  values:
//  - bar1
//  - bar2
//===
//- name: foo
//  values:
//  - bar2
//  - bar1
//```
//
//```
//- name: foo1
//  value: bar1
//- name: foo2
//  value: bar2
//===
//- name: foo2
//  value: bar2
//- name: foo1
//  value: bar1
//```
//
//```
//v1
//  foo1: bar1
//  foo2: bar2
//===
//v2
//- name: foo1
//  value: bar1
//- name: foo2
//  value: bar2
//```

type RunConfigurationSpec struct {
	Pipeline          PipelineIdentifier `json:"pipeline,omitempty"`
	ExperimentName    string             `json:"experimentName,omitempty"`
	Schedule          string             `json:"schedule,omitempty"`
	RuntimeParameters map[string]string  `json:"runtimeParameters,omitempty"`
}

func (rc RunConfiguration) ComputeHash() []byte {
	oh := NewObjectHasher()
	oh.WriteStringField(rc.Spec.Pipeline.String())
	oh.WriteStringField(rc.Spec.ExperimentName)
	oh.WriteStringField(rc.Spec.Schedule)
	oh.WriteStringField(rc.Status.ObservedPipelineVersion)
	oh.WriteMapField(rc.Spec.RuntimeParameters)
	return oh.Sum()
}

func (rc RunConfiguration) ComputeVersion() string {
	hash := rc.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

type RunConfigurationStatus struct {
	Status                  `json:",inline"`
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

func (rc *RunConfiguration) GetStatus() Status {
	return rc.Status.Status
}

func (rc *RunConfiguration) SetStatus(status Status) {
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
