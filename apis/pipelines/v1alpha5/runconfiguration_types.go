package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Triggers struct {
	Schedules         []string       `json:"schedules,omitempty"`
	OnChange          []OnChangeType `json:"onChange,omitempty"`
	RunConfigurations []string       `json:"runConfigurations,omitempty"`
}

// +kubebuilder:validation:Enum=pipeline
type OnChangeType string

var OnChangeTypes = struct {
	Pipeline OnChangeType
}{
	Pipeline: "pipeline",
}

type RunConfigurationSpec struct {
	Run      RunSpec  `json:"run,omitempty"`
	Triggers Triggers `json:"triggers,omitempty"`
}

type TriggeredRunReference struct {
	ProviderId string `json:"providerId,omitempty"`
}

type TriggersStatus struct {
	RunConfigurations map[string]TriggeredRunReference `json:"runConfigurations,omitempty"`
}

type LatestRuns struct {
	Succeeded RunReference `json:"succeeded,omitempty"`
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
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="mlrc"
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider"
//+kubebuilder:storageversion

type RunConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunConfigurationSpec   `json:"spec,omitempty"`
	Status RunConfigurationStatus `json:"status,omitempty"`
}

func (rc *RunConfiguration) SetDependencyRuns(references map[string]RunReference) {
	rc.Status.Dependencies.RunConfigurations = references
}

func (rc *RunConfiguration) GetDependencyRuns() map[string]RunReference {
	if rc.Status.Dependencies.RunConfigurations == nil {
		return make(map[string]RunReference)
	}
	return rc.Status.Dependencies.RunConfigurations
}

func (rc *RunConfiguration) GetRCRuntimeParameters() []RunConfigurationRef {
	return pipelines.Collect(rc.Spec.Run.RuntimeParameters, func(rp RuntimeParameter) (RunConfigurationRef, bool) {
		if rp.ValueFrom == nil {
			return RunConfigurationRef{}, false
		}

		return rp.ValueFrom.RunConfigurationRef, true
	})
}

func (rc *RunConfiguration) GetTriggeringRCs() []RunConfigurationRef {
	return pipelines.Map(rc.Spec.Triggers.RunConfigurations, func(rcName string) RunConfigurationRef {
		return RunConfigurationRef{
			Name: rcName,
		}
	})
}

func (rc *RunConfiguration) GetReferencedRCs() []RunConfigurationRef {
	return append(rc.GetRCRuntimeParameters(), rc.GetTriggeringRCs()...)
}

func (rc *RunConfiguration) GetProvider() string {
	return rc.Status.Provider
}

func (rc *RunConfiguration) GetPipeline() PipelineIdentifier {
	return rc.Spec.Run.Pipeline
}

func (rc *RunConfiguration) GetObservedPipelineVersion() string {
	return rc.Status.ObservedPipelineVersion
}

func (rc *RunConfiguration) SetObservedPipelineVersion(observedPipelineVersion string) {
	rc.Status.ObservedPipelineVersion = observedPipelineVersion
}

func (rc *RunConfiguration) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      rc.Name,
		Namespace: rc.Namespace,
	}
}

func (rc *RunConfiguration) GetKind() string {
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
