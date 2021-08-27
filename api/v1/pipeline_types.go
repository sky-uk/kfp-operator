package v1

import (
	"crypto/sha1"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ComputeVersion(pipelineSpec PipelineSpec) string {
	h := sha1.New()

	h.Write([]byte(pipelineSpec.Image))
	h.Write([]byte(pipelineSpec.TfxComponents))

	keys := make([]string, 0, len(pipelineSpec.Env))
	for k := range pipelineSpec.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		h.Write([]byte(k))
		h.Write([]byte(pipelineSpec.Env[k]))
	}
	version := h.Sum(nil)

	return fmt.Sprintf("%x\n", version)
}

type PipelineSpec struct {
	Image         string            `json:"image" yaml:"image"`
	TfxComponents string            `json:"tfxComponents" yaml:"tfxComponents"`
	Env           map[string]string `json:"env,omitempty" yaml:"env"`
}

type SynchronizationState string

const (
	Unknown   SynchronizationState = ""
	Creating  SynchronizationState = "Creating"
	Succeeded SynchronizationState = "Succeeded"
	Updating  SynchronizationState = "Updating"
	Deleting  SynchronizationState = "Deleting"
	Deleted   SynchronizationState = "Deleted"
	Failed    SynchronizationState = "Failed"
)

type PipelineStatus struct {
	Id                   string               `json:"id,omitempty"`
	Version              string               `json:"version,omitempty"`
	SynchronizationState SynchronizationState `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type Pipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PipelineSpec   `json:"spec,omitempty"`
	Status PipelineStatus `json:"status,omitempty"`
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
