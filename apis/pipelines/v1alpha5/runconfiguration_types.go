package v1alpha5

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"regexp"
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
const ArtifactPathPattern = `^([^\[\]]+)(?:\[([^\[\]]+)\])?$`
// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Pattern=`^([^\[\]]+)(?:\[([^\[\]]+)\])?$`
type ArtifactPath struct {
	Path   string `json:"-" yaml:"-"`
	Filter string `json:"-" yaml:"-"`
}

func (ap ArtifactPath) String() (string, error) {
	if ap.Filter == "" {
		return ap.Path, nil
	}

	if ap.Path == ""  {
		return "", fmt.Errorf("artifact path provided without a path")
	}

	return fmt.Sprintf("%s[%s]", ap.Path, ap.Filter), nil
}

func ArtifactPathFromString(path string) (artifactPath ArtifactPath, err error) {
	pathPattern := regexp.MustCompile(ArtifactPathPattern)
	matches := pathPattern.FindStringSubmatch(path)

	if len(matches) == 0 {
		err = fmt.Errorf("ArtifactPath must match pattern %s", ArtifactPathPattern)
	}

	artifactPath.Path = matches[0]

	if len(matches) > 1 {
		artifactPath.Filter = matches[1]
	}

	return
}

func (ap ArtifactPath) MarshalText() ([]byte, error) {
	serialised, err := ap.String()
	if err != nil {
		return nil, err
	}

	return []byte(serialised), nil
}

func (ap *ArtifactPath) UnmarshalText(bytes []byte) error {
	deserialised, err := ArtifactPathFromString(string(bytes))
	*ap = deserialised

	return err
}

type Artifact struct {
	Name string       `json:"name"`
	Path []ArtifactPath `json:"path"`
}

type RunConfigurationSpec struct {
	Run      RunSpec  `json:"run,omitempty"`
	Triggers Triggers `json:"triggers,omitempty"`
	Artifacts []Artifact `json:"artifacts,omitempty"`
}

type RunReference struct {
	ProviderId string `json:"providerId,omitempty"`
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
