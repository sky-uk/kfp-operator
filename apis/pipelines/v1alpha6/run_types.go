package v1alpha6

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RuntimeParameter struct {
	Name      string     `json:"name"`
	Value     string     `json:"value,omitempty"`
	ValueFrom *ValueFrom `json:"valueFrom,omitempty"`
}

type RunConfigurationRef struct {
	Name           string `json:"name"`
	OutputArtifact string `json:"outputArtifact"`
}

type ValueFrom struct {
	RunConfigurationRef RunConfigurationRef `json:"runConfigurationRef"`
}

type RunSpec struct {
	Provider          string             `json:"provider" yaml:"provider"`
	Pipeline          PipelineIdentifier `json:"pipeline,omitempty"`
	ExperimentName    string             `json:"experimentName,omitempty"`
	RuntimeParameters []RuntimeParameter `json:"runtimeParameters,omitempty"`
	// Needed for conversion *only*
	// +kubebuilder:validation:-
	// +optional
	Parameters []RuntimeParameter `json:"parameters,omitempty"`
	Artifacts  []OutputArtifact   `json:"artifacts,omitempty"`
}

func (runSpec *RunSpec) ResolveRuntimeParameters(dependencies Dependencies) ([]apis.NamedValue, error) {
	return apis.MapErr(runSpec.RuntimeParameters, func(r RuntimeParameter) (apis.NamedValue, error) {
		if r.ValueFrom == nil {
			return apis.NamedValue{
				Name:  r.Name,
				Value: r.Value,
			}, nil
		}

		if dependency, ok := dependencies.RunConfigurations[r.ValueFrom.RunConfigurationRef.Name]; ok {
			for _, artifact := range dependency.Artifacts {
				if artifact.Name == r.ValueFrom.RunConfigurationRef.OutputArtifact {
					return apis.NamedValue{
						Name:  r.Name,
						Value: artifact.Location,
					}, nil
				}
			}

			return apis.NamedValue{}, fmt.Errorf("artifact '%s' not found in dependency '%s'", r.ValueFrom.RunConfigurationRef.OutputArtifact, r.ValueFrom.RunConfigurationRef.Name)
		}

		return apis.NamedValue{}, fmt.Errorf("dependency '%s' not found", r.ValueFrom.RunConfigurationRef.Name)
	})
}

func (runSpec *RunSpec) HasUnmetDependencies(dependencies Dependencies) bool {
	_, err := runSpec.ResolveRuntimeParameters(dependencies)
	return err != nil
}

func cmpRuntimeParameters(rp1, rp2 RuntimeParameter) bool {
	if rp1.Name != rp2.Name {
		return rp1.Name < rp2.Name
	}

	if rp1.Value != rp2.Value {
		return rp1.Value < rp2.Value
	}

	if rp1.ValueFrom == nil {
		return rp2.ValueFrom != nil
	}

	if rp1.ValueFrom.RunConfigurationRef.Name != rp2.ValueFrom.RunConfigurationRef.Name {
		return rp1.ValueFrom.RunConfigurationRef.Name < rp2.ValueFrom.RunConfigurationRef.Name
	}

	return rp1.ValueFrom.RunConfigurationRef.OutputArtifact < rp2.ValueFrom.RunConfigurationRef.OutputArtifact
}

func writeRuntimeParameter(oh pipelines.ObjectHasher, rp RuntimeParameter) {
	oh.WriteStringField(rp.Name)
	oh.WriteStringField(rp.Value)
	if rp.ValueFrom != nil {
		oh.WriteStringField(rp.ValueFrom.RunConfigurationRef.Name)
		oh.WriteStringField(rp.ValueFrom.RunConfigurationRef.OutputArtifact)
	}
}

func WriteRunTimeParameters(oh pipelines.ObjectHasher, rps []RuntimeParameter) {
	pipelines.WriteList(oh, rps, cmpRuntimeParameters, writeRuntimeParameter)
}

func (rs RunSpec) WriteRunSpec(oh pipelines.ObjectHasher) {
	oh.WriteStringField(rs.Pipeline.String())
	oh.WriteStringField(rs.ExperimentName)
	WriteRunTimeParameters(oh, rs.RuntimeParameters)
	pipelines.WriteKVListField(oh, rs.Artifacts)
}

func (rs RunSpec) ComputeVersion() string {
	oh := pipelines.NewObjectHasher()
	rs.WriteRunSpec(oh)
	hash := oh.Sum()[0:3]

	return fmt.Sprintf("%x", hash)
}

func (r Run) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	r.Spec.WriteRunSpec(oh)
	oh.WriteStringField(r.Status.ObservedPipelineVersion)
	return oh.Sum()
}

func (r Run) ComputeVersion() string {
	hash := r.ComputeHash()[0:3]

	return fmt.Sprintf("%x", hash)
}

type CompletionState string

var CompletionStates = struct {
	Succeeded CompletionState
	Failed    CompletionState
}{
	Succeeded: "Succeeded",
	Failed:    "Failed",
}

type RunReference struct {
	ProviderId string            `json:"providerId,omitempty"`
	Artifacts  []common.Artifact `json:"artifacts,omitempty"`
}

type Dependencies struct {
	RunConfigurations map[string]RunReference `json:"runConfigurations,omitempty"`
}

type RunStatus struct {
	Status                  `json:",inline"`
	ObservedPipelineVersion string          `json:"observedPipelineVersion,omitempty"`
	Dependencies            Dependencies    `json:"dependencies,omitempty"`
	CompletionState         CompletionState `json:"completionState,omitempty"`
	MarkedCompletedAt       *metav1.Time    `json:"markedCompletedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="mlr"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider.name"
// +kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.conditions[?(@.type==\"Synchronized\")].reason"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
// +kubebuilder:printcolumn:name="CompletionState",type="string",JSONPath=".status.completionState"
type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunSpec   `json:"spec,omitempty"`
	Status RunStatus `json:"status,omitempty"`
}

func (r *Run) SetDependencyRuns(references map[string]RunReference) {
	r.Status.Dependencies.RunConfigurations = references
}

func (r *Run) GetDependencyRuns() map[string]RunReference {
	if r.Status.Dependencies.RunConfigurations == nil {
		return make(map[string]RunReference)
	}
	return r.Status.Dependencies.RunConfigurations
}

func (r *Run) GetReferencedRCArtifacts() []RunConfigurationRef {
	return apis.Collect(r.Spec.RuntimeParameters, func(rp RuntimeParameter) (RunConfigurationRef, bool) {
		if rp.ValueFrom == nil {
			return RunConfigurationRef{}, false
		}

		return rp.ValueFrom.RunConfigurationRef, true
	})
}

func (r *Run) GetReferencedRCs() []string {
	return apis.Collect(r.Spec.RuntimeParameters, func(rp RuntimeParameter) (string, bool) {
		if rp.ValueFrom == nil {
			return "", false
		}

		return rp.ValueFrom.RunConfigurationRef.Name, true
	})
}

func (r *Run) GetPipeline() PipelineIdentifier {
	return r.Spec.Pipeline
}

func (r *Run) GetObservedPipelineVersion() string {
	return r.Status.ObservedPipelineVersion
}

func (r *Run) SetObservedPipelineVersion(newVersion string) {
	r.Status.ObservedPipelineVersion = newVersion
}

func (r *Run) GetStatus() Status {
	return r.Status.Status
}

func (r *Run) SetStatus(status Status) {
	r.Status.Status = status
}

func (r *Run) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      r.Name,
		Namespace: r.Namespace,
	}
}

func (r *Run) GetKind() string {
	return "run"
}

// +kubebuilder:object:root=true
type RunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Run `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Run{}, &RunList{})
}
