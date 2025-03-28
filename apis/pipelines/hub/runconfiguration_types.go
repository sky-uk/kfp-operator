package v1beta1

import (
	"reflect"

	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Triggers struct {
	Schedules         []Schedule     `json:"schedules,omitempty"`
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
	SynchronizationState     apis.SynchronizationState `json:"synchronizationState,omitempty"`
	Provider                 common.NamespacedName     `json:"provider,omitempty"`
	ObservedPipelineVersion  string                    `json:"observedPipelineVersion,omitempty"`
	TriggeredPipelineVersion string                    `json:"triggeredPipelineVersion,omitempty"`
	LatestRuns               LatestRuns                `json:"latestRuns,omitempty"`
	Dependencies             Dependencies              `json:"dependencies,omitempty"`
	Triggers                 TriggersStatus            `json:"triggers,omitempty"`
	ObservedGeneration       int64                     `json:"observedGeneration,omitempty"`
	Conditions               Conditions                `json:"conditions,omitempty"`
}

func (rcs *RunConfigurationStatus) SetSynchronizationState(state apis.SynchronizationState, message string) {
	rcs.SynchronizationState = state
	condition := metav1.Condition{
		Type:               ConditionTypes.SynchronizationSucceeded,
		Message:            message,
		ObservedGeneration: rcs.ObservedGeneration,
		Reason:             string(state),
		LastTransitionTime: metav1.Now(),
		Status:             ConditionStatusForSynchronizationState(state),
	}
	rcs.Conditions = rcs.Conditions.MergeIntoConditions(condition)
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="mlrc"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider"
// +kubebuilder:storageversion
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
