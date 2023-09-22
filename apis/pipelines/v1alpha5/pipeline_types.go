package v1alpha5

import (
	"fmt"
	. "github.com/docker/distribution/reference"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (ps Pipeline) ComputeHash() []byte {
	oh := pipelines.NewObjectHasher()
	oh.WriteStringField(ps.Spec.Image)
	oh.WriteStringField(ps.Spec.TfxComponents)
	pipelines.WriteKVListField(oh, ps.Spec.Env)
	pipelines.WriteKVListField(oh, ps.Spec.BeamArgs)
	return oh.Sum()
}

func (ps Pipeline) ComputeVersion() string {
	hash := ps.ComputeHash()[0:3]
	ref, err := ParseNormalizedNamed(ps.Spec.Image)

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
//+kubebuilder:printcolumn:name="ProviderId",type="string",JSONPath=".status.providerId"
//+kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.synchronizationState"
//+kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version"

type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   hub.PipelineSpec `json:"spec,omitempty"`
	Status Status           `json:"status,omitempty"`
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
