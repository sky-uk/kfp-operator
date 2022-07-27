package v1alpha2

import (
	"fmt"
	. "github.com/docker/distribution/reference"
	"github.com/sky-uk/kfp-operator/controllers/objecthasher"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type PipelineSpec struct {
	Image         string            `json:"image" yaml:"image"`
	TfxComponents string            `json:"tfxComponents" yaml:"tfxComponents"`
	Env           map[string]string `json:"env,omitempty" yaml:"env"`
	BeamArgs      map[string]string `json:"beamArgs,omitempty" yaml:"beamArgs"`
}

func (ps PipelineSpec) ComputeHash() []byte {
	oh := objecthasher.New()
	oh.WriteStringField(ps.Image)
	oh.WriteStringField(ps.TfxComponents)
	oh.WriteMapField(ps.Env)
	oh.WriteMapField(ps.BeamArgs)
	return oh.Sum()
}

func (ps PipelineSpec) ComputeVersion() string {
	hash := ps.ComputeHash()[0:3]
	ref, err := ParseNormalizedNamed(ps.Image)

	if err == nil {
		if namedTagged, ok := TagNameOnly(ref).(NamedTagged); ok {
			return fmt.Sprintf("%s-%x", namedTagged.Tag(), hash)
		}
	}
	return fmt.Sprintf("%x", hash)
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:shortName="mlp"
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="KfpId",type="string",JSONPath=".status.kfpId"
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"

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

func init() {
	SchemeBuilder.Register(&Pipeline{}, &PipelineList{})
}
