package v1beta1

import (
	"reflect"

	"github.com/samber/lo"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Triggers struct {
	Schedules         []Schedule              `json:"schedules,omitempty"`
	OnChange          []OnChangeType          `json:"onChange,omitempty"`
	RunConfigurations []common.NamespacedName `json:"runConfigurations,omitempty"`
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

type PipelineTriggerStatus struct {
	Version string `json:"version,omitempty"`
}

type TriggersStatus struct {
	RunConfigurations map[string]TriggeredRunReference `json:"runConfigurations,omitempty"`
	RunSpec           RunSpecTriggerStatus             `json:"runSpec,omitempty"`
	Pipeline          PipelineTriggerStatus            `json:"pipeline,omitempty"`
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
	Provider           common.NamespacedName `json:"provider,omitempty"`
	LatestRuns         LatestRuns            `json:"latestRuns,omitempty"`
	Dependencies       Dependencies          `json:"dependencies,omitempty"`
	Triggers           TriggersStatus        `json:"triggers,omitempty"`
	ObservedGeneration int64                 `json:"observedGeneration,omitempty"`
	Conditions         apis.Conditions       `json:"conditions,omitempty"`
}

func (rcs *RunConfigurationStatus) SetSynchronizationState(state apis.SynchronizationState, message string) {
	condition := metav1.Condition{
		Type:               apis.ConditionTypes.SynchronizationSucceeded,
		Message:            message,
		ObservedGeneration: rcs.ObservedGeneration,
		Reason:             string(state),
		LastTransitionTime: metav1.Now(),
		Status:             apis.ConditionStatusForSynchronizationState(state),
	}
	rcs.Conditions = rcs.Conditions.MergeIntoConditions(condition)
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="mlrc"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.conditions[?(@.type==\"Synchronized\")].reason"
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
	return apis.Collect(rc.Spec.Run.Parameters, func(p Parameter) (RunConfigurationRef, bool) {
		if p.ValueFrom == nil {
			return RunConfigurationRef{}, false
		}

		return p.ValueFrom.RunConfigurationRef, true
	})
}

func (rc *RunConfiguration) GetReferencedRCs() []common.NamespacedName {
	triggeringRcs := lo.Map(rc.Spec.Triggers.RunConfigurations, func(rcName common.NamespacedName, _ int) common.NamespacedName {
		return rcName
	})

	parameterRcs := lo.Map(rc.GetReferencedRCArtifacts(), func(r RunConfigurationRef, _ int) common.NamespacedName {
		return r.Name
	})

	return lo.Uniq(append(parameterRcs, triggeringRcs...))
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
