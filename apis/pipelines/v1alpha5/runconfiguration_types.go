package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
)

type TriggersStatus struct {
	RunConfigurations map[string]hub.TriggeredRunReference `json:"runConfigurations,omitempty"`
	RunSpec           hub.RunSpecTriggerStatus             `json:"runSpec,omitempty"`
}

func (ts TriggersStatus) Equals(other TriggersStatus) bool {
	if ts.RunSpec.Version != other.RunSpec.Version {
		return false
	}

	if len(ts.RunConfigurations) == 0 && len(other.RunConfigurations) == 0 {
		return true
	}

	return reflect.DeepEqual(ts.RunConfigurations, other.RunConfigurations)
}

type LatestRuns struct {
	Succeeded hub.RunReference `json:"succeeded,omitempty"`
}

type RunConfigurationStatus struct {
	SynchronizationState     apis.SynchronizationState `json:"synchronizationState,omitempty"`
	Provider                 string                    `json:"provider,omitempty"`
	ObservedPipelineVersion  string                    `json:"observedPipelineVersion,omitempty"`
	TriggeredPipelineVersion string                    `json:"triggeredPipelineVersion,omitempty"`
	LatestRuns               LatestRuns                `json:"latestRuns,omitempty"`
	Dependencies             Dependencies              `json:"dependencies,omitempty"`
	Triggers                 TriggersStatus            `json:"triggers,omitempty"`
	ObservedGeneration       int64                     `json:"observedGeneration,omitempty"`
	Conditions               hub.Conditions            `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="mlrc"
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider"

type RunConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   hub.RunConfigurationSpec `json:"spec,omitempty"`
	Status RunConfigurationStatus   `json:"status,omitempty"`
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
