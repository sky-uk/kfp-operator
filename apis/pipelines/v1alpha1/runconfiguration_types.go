package v1alpha1

import (
	"fmt"
	"strings"

	"github.com/sky-uk/kfp-operator/controllers/objecthasher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RunConfigurationSpec struct {
	PipelineName      string            `json:"pipelineName,omitempty"` // TODO: Remove PipelineName
	Pipeline          string            `json:"pipeline,omitempty"`
	ExperimentName    string            `json:"experimentName,omitempty"`
	Schedule          string            `json:"schedule,omitempty"`
	RuntimeParameters map[string]string `json:"runtimeParameters,omitempty"`
}

func (rcs RunConfiguration) ComputeHash() []byte {
	oh := objecthasher.New()
	oh.WriteStringField(rcs.Spec.Pipeline)
	oh.WriteStringField(rcs.Spec.ExperimentName)
	oh.WriteStringField(rcs.Spec.Schedule)
	oh.WriteMapField(rcs.Spec.RuntimeParameters)
	oh.WriteStringField(rcs.Status.ObservedPipelineVersion)
	return oh.Sum()
}

func (rcs RunConfiguration) ComputeVersion() string {
	hash := rcs.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

func (rcs RunConfiguration) ExtractPipelineNameVersion() (string, string) {
	if rcs.Spec.Pipeline == "" {
		return "", ""
	}

	nameVersion := strings.Split(rcs.Spec.Pipeline, ":")

	if len(nameVersion) < 2 {
		return nameVersion[0], ""
	}

	return nameVersion[0], nameVersion[1]
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
