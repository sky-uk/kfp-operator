package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Triggers struct {
	Schedules []string       `json:"schedules,omitempty"`
	OnChange  []OnChangeType `json:"onChange,omitempty"`
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

func (rc *RunConfiguration) SetDependencyRun(name string, reference RunReference) {
	if rc.Status.Dependencies.RunConfigurations == nil {
		rc.Status.Dependencies.RunConfigurations = make(map[string]RunReference, 1)
	}

	rc.Status.Dependencies.RunConfigurations[name] = reference
}

func (rc *RunConfiguration) GetReferencedDependencies() []string {
	return pipelines.Collect(rc.Spec.Run.RuntimeParameters, func(rp RuntimeParameter) (string, bool) {
		if rp.ValueFrom == nil {
			return "", false
		}

		return rp.ValueFrom.RunConfigurationRef.Name, true
	})
}

func (rc *RunConfiguration) GetDependencyRun(name string) (RunReference, bool) {
	ref, ok := rc.Status.Dependencies.RunConfigurations[name]
	return ref, ok
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
