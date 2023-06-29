package v1alpha5

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RuntimeParameter struct {
	Name      string    `json:"Name"`
	Value     string    `json:"value,omitempty"`
	ValueFrom ValueFrom `json:"valueFrom,omitempty"`
}

type RunConfigurationRef struct {
	Name string           `json:"Name"`
	OutputArtifact string `json:"outputArtifact"`
}

type ValueFrom struct {
	RunConfigurationRef RunConfigurationRef `json:"runConfigurationRef"`
}

func (v RuntimeParameter) GetKey() string {
	return v.Name
}

func (v RuntimeParameter) GetValue() string {
	if v.Value != "" {
		return v.Value
	} else {
		return v.ValueFrom.RunConfigurationRef.Name+v.ValueFrom.RunConfigurationRef.OutputArtifact
	}
}

type RunSpec struct {
	Pipeline          PipelineIdentifier `json:"pipeline,omitempty"`
	ExperimentName    string             `json:"experimentName,omitempty"`
	RuntimeParameters []RuntimeParameter     `json:"runtimeParameters,omitempty"`
	Artifacts         []OutputArtifact   `json:"artifacts,omitempty"`
}

func (r Run) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(r.Spec.Pipeline.String())
	oh.WriteStringField(r.Spec.ExperimentName)
	pipelines.WriteKVListField(oh, r.Spec.RuntimeParameters)
	pipelines.WriteKVListField(oh, r.Spec.Artifacts)
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

type RunStatus struct {
	Status                  `json:",inline"`
	ObservedPipelineVersion string          `json:"observedPipelineVersion,omitempty"`
	Dependencies            map[string]RunReference  `json:"dependencies,omitempty"`
	CompletionState         CompletionState `json:"completionState,omitempty"`
	MarkedCompletedAt       *metav1.Time    `json:"markedCompletedAt,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="mlr"
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="ProviderId",type="string",JSONPath=".status.providerId"
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
//+kubebuilder:printcolumn:name="CompletionState",type="string",JSONPath=".status.completionState"
//+kubebuilder:storageversion

type Run struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RunSpec   `json:"spec,omitempty"`
	Status RunStatus `json:"status,omitempty"`
}

func (rc *Run) SetDependency(name string, reference RunReference) {
	if rc.Status.Dependencies == nil {
		rc.Status.Dependencies = make(map[string]RunReference, 1)
	}

	rc.Status.Dependencies[name] = reference
}

func (rc *Run) GetDependencies() map[string]RunReference {
	if rc.Status.Dependencies != nil {
		return rc.Status.Dependencies
	} else {
		return make(map[string]RunReference, 1)
	}
}

func (rc *Run) GetRunConfigurations() []string {
	return apis.Collect(rc.Spec.RuntimeParameters, func(rp RuntimeParameter) (string, bool) {
		rc := rp.ValueFrom.RunConfigurationRef.Name
		return rc, rc != ""
	})
}

func (r *Run) GetProvider() string {
	return r.Status.ProviderId.Provider
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

//+kubebuilder:object:root=true

type RunList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Run `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Run{}, &RunList{})
}
