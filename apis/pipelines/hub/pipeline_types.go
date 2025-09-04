package v1beta1

import (
	"encoding/json"
	"fmt"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"strings"

	"github.com/distribution/reference"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type PipelineSpec struct {
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?/[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Provider  common.NamespacedName `json:"provider" yaml:"provider"`
	Image     string                `json:"image" yaml:"image"`
	Env       []apis.NamedValue     `json:"env,omitempty" yaml:"env"`
	Framework PipelineFramework     `json:"framework" yaml:"framework"`
}

type PipelineFramework struct {
	Name       string                           `json:"name" yaml:"name"`
	Parameters map[string]*apiextensionsv1.JSON `json:"parameters" yaml:"parameters"`
}

const FallbackFramework = "tfx"

func NewPipelineFramework(compilerType string) PipelineFramework {
	return PipelineFramework{
		Name:       compilerType,
		Parameters: make(map[string]*apiextensionsv1.JSON),
	}
}

func (ps Pipeline) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(ps.Spec.Image)
	pipelines.WriteKVListField(oh, ps.Spec.Env)
	oh.WriteStringField(ps.Spec.Framework.Name)
	oh.WriteJSONMapField(ps.Spec.Framework.Parameters)

	return oh.Sum()
}

func (ps Pipeline) ComputeVersion() string {
	computeHash := ps.ComputeHash()
	if computeHash == nil {
		return ""
	}
	hash := computeHash[0:3]
	ref, err := reference.ParseNormalizedNamed(ps.Spec.Image)

	if err == nil {
		if namedTagged, ok := reference.TagNameOnly(ref).(reference.NamedTagged); ok {
			return fmt.Sprintf("%s-%x", namedTagged.Tag(), hash)
		}
	}
	return fmt.Sprintf("%x", hash)
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="mlp"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Provider",type="string",JSONPath=".status.provider.name"
// +kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.conditions[?(@.type==\"Synchronized\")].reason"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"
// +kubebuilder:storageversion
type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec `json:"spec,omitempty"`
	Status Status       `json:"status,omitempty"`
}

func (p *Pipeline) GetStatus() Status {
	return p.Status
}

func (p *Pipeline) SetStatus(status Status) {
	p.Status = status
}

func (p Pipeline) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      p.Name,
		Namespace: p.Namespace,
	}
}

func (p Pipeline) GetKind() string {
	return "pipeline"
}

//+kubebuilder:object:root=true

type PipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Pipeline `json:"items"`
}

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Pattern:=`^[\w\-]+(?::[\w\-_.]+)?$`
type PipelineIdentifier struct {
	Name    string `json:"-"`
	Version string `json:"-"`
}

func (pid *PipelineIdentifier) String() string {
	if pid.Version == "" {
		return pid.Name
	}

	return strings.Join([]string{pid.Name, pid.Version}, ":")
}

func (pid *PipelineIdentifier) MarshalJSON() ([]byte, error) {
	return json.Marshal(pid.String())
}

func (pid *PipelineIdentifier) UnmarshalJSON(bytes []byte) error {
	var pidStr string
	err := json.Unmarshal(bytes, &pidStr)
	if err != nil {
		return err
	}

	nameVersion := strings.Split(pidStr, ":")
	pid.Name = nameVersion[0]

	if len(nameVersion) == 2 {
		pid.Version = nameVersion[1]
	}

	return nil
}

func (pipeline *Pipeline) UnversionedIdentifier() PipelineIdentifier {
	return PipelineIdentifier{Name: pipeline.Name}
}

func (pipeline *Pipeline) VersionedIdentifier() PipelineIdentifier {
	return PipelineIdentifier{Name: pipeline.Name, Version: pipeline.ComputeVersion()}
}

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
