package v1alpha5

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type RuntimeParameter struct {
	Name      string    `json:"name"`
	Value     string    `json:"value,omitempty"`
	ValueFrom ValueFrom `json:"valueFrom,omitempty"`
}

type RunConfigurationRef struct {
	Name           string `json:"name"`
	OutputArtifact string `json:"outputArtifact"`
}

type ValueFrom struct {
	RunConfigurationRef RunConfigurationRef `json:"runConfigurationRef"`
}

type RunSpec struct {
	Pipeline          PipelineIdentifier `json:"pipeline,omitempty"`
	ExperimentName    string             `json:"experimentName,omitempty"`
	RuntimeParameters []RuntimeParameter `json:"runtimeParameters,omitempty"`
	Artifacts         []OutputArtifact   `json:"artifacts,omitempty"`
}

func WriteRunTimeParameters(oh pipelines.ObjectHasher, rts []RuntimeParameter) {
	pipelines.WriteList(oh, rts, func(rt1, rt2 RuntimeParameter) bool {
		if rt1.Name != rt2.Name {
			return rt1.Name < rt2.Name
		} else if rt1.Value != rt2.Value {
			return rt1.Value < rt2.Value
		} else if rt1.ValueFrom.RunConfigurationRef.Name != rt1.ValueFrom.RunConfigurationRef.Name {
			return rt1.ValueFrom.RunConfigurationRef.Name < rt1.ValueFrom.RunConfigurationRef.Name
		} else {
			return rt1.ValueFrom.RunConfigurationRef.OutputArtifact < rt1.ValueFrom.RunConfigurationRef.OutputArtifact
		}
	}, func(oh pipelines.ObjectHasher, rt RuntimeParameter) {
		oh.WriteStringField(rt.Name)
		oh.WriteStringField(rt.Value)
		oh.WriteStringField(rt.ValueFrom.RunConfigurationRef.Name)
		oh.WriteStringField(rt.ValueFrom.RunConfigurationRef.OutputArtifact)
	})
}

func (r Run) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(r.Spec.Pipeline.String())
	oh.WriteStringField(r.Spec.ExperimentName)
	WriteRunTimeParameters(oh, r.Spec.RuntimeParameters)
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
	ObservedPipelineVersion string                  `json:"observedPipelineVersion,omitempty"`
	Dependencies            map[string]RunReference `json:"dependencies,omitempty"`
	CompletionState         CompletionState         `json:"completionState,omitempty"`
	MarkedCompletedAt       *metav1.Time            `json:"markedCompletedAt,omitempty"`
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
	return pipelines.Collect(rc.Spec.RuntimeParameters, func(rp RuntimeParameter) (string, bool) {
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
