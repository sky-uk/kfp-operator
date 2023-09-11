package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

type Triggers struct {
	Schedules         []string       `json:"schedules,omitempty"`
	OnChange          []OnChangeType `json:"onChange,omitempty"`
	RunConfigurations []string       `json:"runConfigurations,omitempty"`
}

// +kubebuilder:validation:Enum=pipeline;runSpec
type OnChangeType string

var OnChangeTypes = struct {
	Pipeline OnChangeType
	RunSpec  OnChangeType
}{
	Pipeline: "pipeline",
	RunSpec:  "runSpec",
}

type RunConfigurationSpec struct {
	Run      RunSpec  `json:"run,omitempty"`
	Triggers Triggers `json:"triggers,omitempty"`
}

type TriggeredRunReference struct {
	ProviderId string `json:"providerId,omitempty"`
}

type RunSpecTriggerStatus struct {
	Version string `json:"version,omitempty"`
}

type TriggersStatus struct {
	RunConfigurations map[string]TriggeredRunReference `json:"runConfigurations,omitempty"`
	RunSpec           RunSpecTriggerStatus             `json:"runSpec,omitempty"`
	Pipeline          PipelineReference                `json:"pipeline,omitempty"`
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
	Succeeded RunReference `json:"succeeded,omitempty"`
}

type RunConfigurationStatus struct {
	SynchronizationState apis.SynchronizationState `json:"synchronizationState,omitempty"`
	Provider             string                    `json:"provider,omitempty"`
	LatestRuns           LatestRuns                `json:"latestRuns,omitempty"`
	Dependencies         Dependencies              `json:"dependencies,omitempty"`
	Triggers             TriggersStatus            `json:"triggers,omitempty"`
	ObservedGeneration   int64                     `json:"observedGeneration,omitempty"`
	Conditions           Conditions                `json:"conditions,omitempty"`
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

func (rc *RunConfiguration) GetReferencedRCArtifacts() []RunConfigurationRef {
	return pipelines.Collect(rc.Spec.Run.RuntimeParameters, func(rp RuntimeParameter) (RunConfigurationRef, bool) {
		if rp.ValueFrom == nil {
			return RunConfigurationRef{}, false
		}

		return rp.ValueFrom.RunConfigurationRef, true
	})
}

func (rc *RunConfiguration) GetReferencedRCs() []string {
	triggeringRcs := pipelines.Map(rc.Spec.Triggers.RunConfigurations, func(rcName string) string {
		return rcName
	})

	parameterRcs := pipelines.Map(rc.GetReferencedRCArtifacts(), func(r RunConfigurationRef) string {
		return r.Name
	})

	return pipelines.Unique(append(parameterRcs, triggeringRcs...))
}

func (rc *RunConfiguration) GetProvider() string {
	return rc.Status.Provider
}

func (rc *RunConfiguration) GetPipeline() PipelineIdentifier {
	return rc.Spec.Run.Pipeline
}

func (rc *RunConfiguration) GetObservedPipelineVersion() string {
	return rc.Status.Dependencies.Pipeline.Version
}

func (rc *RunConfiguration) SetObservedPipelineVersion(observedPipelineVersion string) {
	rc.Status.Dependencies.Pipeline.Version = observedPipelineVersion
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
