package v1beta1

import (
	"fmt"
	"strconv"

	"github.com/samber/lo"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type Parameter struct {
	Name      string     `json:"name"`
	Value     string     `json:"value,omitempty"`
	ValueFrom *ValueFrom `json:"valueFrom,omitempty"`
}

type RunConfigurationRef struct {
	Name           common.NamespacedName `json:"name"`
	OutputArtifact string                `json:"outputArtifact"`
	Optional       bool                  `json:"optional,omitempty"`
}

type ValueFrom struct {
	RunConfigurationRef RunConfigurationRef `json:"runConfigurationRef"`
}

type RunSpec struct {
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?/[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Provider       common.NamespacedName `json:"provider" yaml:"provider"`
	Pipeline       PipelineIdentifier    `json:"pipeline,omitempty"`
	ExperimentName string                `json:"experimentName,omitempty"`
	Parameters     []Parameter           `json:"parameters,omitempty"`
	// Deprecated: Needed for conversion only
	// +kubebuilder:validation:-
	// +optional
	RuntimeParameters []Parameter      `json:"runtimeParameters,omitempty"`
	Artifacts         []OutputArtifact `json:"artifacts,omitempty"`
}

func (runSpec *RunSpec) ResolveParameters(dependencies Dependencies) ([]apis.NamedValue, []Parameter, error) {
	unresolvedOptionalParameters := []Parameter{}
	resolvedParameters, err := apis.MapErr(runSpec.Parameters, func(p Parameter) (apis.NamedValue, error) {
		if p.ValueFrom == nil {
			return apis.NamedValue{
				Name:  p.Name,
				Value: p.Value,
			}, nil
		}

		rcNamespacedName, err := p.ValueFrom.RunConfigurationRef.Name.String()
		if err != nil {
			return apis.NamedValue{}, err
		}

		if dependency, ok := dependencies.RunConfigurations[rcNamespacedName]; ok {
			for _, artifact := range dependency.Artifacts {
				if artifact.Name == p.ValueFrom.RunConfigurationRef.OutputArtifact {
					return apis.NamedValue{
						Name:  p.Name,
						Value: artifact.Location,
					}, nil
				}
			}

			if p.ValueFrom.RunConfigurationRef.Optional {
				unresolvedOptionalParameters = append(unresolvedOptionalParameters, p)
				return apis.NamedValue{}, nil
			}

			return apis.NamedValue{}, fmt.Errorf("artifact '%s' not found in dependency '%s'", p.ValueFrom.RunConfigurationRef.OutputArtifact, p.ValueFrom.RunConfigurationRef.Name)
		}

		return apis.NamedValue{}, fmt.Errorf("dependency '%s' not found", p.ValueFrom.RunConfigurationRef.Name)
	})
	return resolvedParameters, unresolvedOptionalParameters, err
}

func (runSpec *RunSpec) HasUnmetDependencies(dependencies Dependencies) bool {
	_, _, err := runSpec.ResolveParameters(dependencies)
	return err != nil
}

func cmpParameters(p1, p2 Parameter) bool {
	if p1.Name != p2.Name {
		return p1.Name < p2.Name
	}

	if p1.Value != p2.Value {
		return p1.Value < p2.Value
	}

	if p1.ValueFrom == nil {
		return p2.ValueFrom != nil
	}

	if p1.ValueFrom.RunConfigurationRef.Name != p2.ValueFrom.RunConfigurationRef.Name {
		return p1.ValueFrom.RunConfigurationRef.Name.Name < p2.ValueFrom.RunConfigurationRef.Name.Name
	}

	if p1.ValueFrom.RunConfigurationRef.OutputArtifact != p2.ValueFrom.RunConfigurationRef.OutputArtifact {
		return p1.ValueFrom.RunConfigurationRef.OutputArtifact < p2.ValueFrom.RunConfigurationRef.OutputArtifact
	}

	return !p1.ValueFrom.RunConfigurationRef.Optional && p1.ValueFrom.RunConfigurationRef.Optional != p2.ValueFrom.RunConfigurationRef.Optional
}

func writeParameter(oh pipelines.ObjectHasher, p Parameter) {
	oh.WriteStringField(p.Name)
	oh.WriteStringField(p.Value)
	if p.ValueFrom != nil {
		oh.WriteStringField(p.ValueFrom.RunConfigurationRef.Name.Name)
		oh.WriteStringField(p.ValueFrom.RunConfigurationRef.Name.Namespace)
		oh.WriteStringField(p.ValueFrom.RunConfigurationRef.OutputArtifact)
		oh.WriteStringField(strconv.FormatBool(p.ValueFrom.RunConfigurationRef.Optional))
	}
}

func WriteParameters(oh pipelines.ObjectHasher, ps []Parameter) {
	pipelines.WriteList(oh, ps, cmpParameters, writeParameter)
}

func (rs RunSpec) WriteRunSpec(oh pipelines.ObjectHasher) {
	oh.WriteStringField(rs.Pipeline.String())
	oh.WriteStringField(rs.ExperimentName)
	WriteParameters(oh, rs.Parameters)
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
	oh.WriteStringField(r.Status.Dependencies.Pipeline.Version)
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

type ObservedPipeline struct {
	Version string `json:"version,omitempty"`
}

type Dependencies struct {
	RunConfigurations map[string]RunReference `json:"runConfigurations,omitempty"`
	Pipeline          ObservedPipeline        `json:"pipeline,omitempty"`
}

type RunStatus struct {
	Status            `json:",inline"`
	Dependencies      Dependencies    `json:"dependencies,omitempty"`
	CompletionState   CompletionState `json:"completionState,omitempty"`
	MarkedCompletedAt *metav1.Time    `json:"markedCompletedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="mlr"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider.name"
// +kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.conditions[?(@.type==\"Synchronized\")].reason"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
// +kubebuilder:printcolumn:name="CompletionState",type="string",JSONPath=".status.completionState"
// +kubebuilder:storageversion
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
	return lo.FilterMap(r.Spec.Parameters, func(p Parameter, _ int) (RunConfigurationRef, bool) {
		if p.ValueFrom == nil {
			return RunConfigurationRef{}, false
		}

		return p.ValueFrom.RunConfigurationRef, true
	})
}

func (r *Run) GetReferencedRCs() []common.NamespacedName {
	return lo.FilterMap(r.Spec.Parameters, func(p Parameter, _ int) (common.NamespacedName, bool) {
		if p.ValueFrom == nil {
			return common.NamespacedName{}, false
		}

		return p.ValueFrom.RunConfigurationRef.Name, true
	})
}

func (r *Run) GetPipeline() PipelineIdentifier {
	return r.Spec.Pipeline
}

func (r *Run) GetObservedPipelineVersion() string {
	return r.Status.Dependencies.Pipeline.Version
}

func (r *Run) SetObservedPipelineVersion(newVersion string) {
	r.Status.Dependencies.Pipeline.Version = newVersion
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
