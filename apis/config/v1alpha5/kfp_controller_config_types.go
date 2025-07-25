package v1alpha5

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type KfpControllerConfigSpec struct {
	DefaultProvider        string               `json:"defaultProvider,omitempty"`
	WorkflowTemplatePrefix string               `json:"workflowTemplatePrefix,omitempty"`
	WorkflowNamespace      string               `json:"workflowNamespace,omitempty"`
	Multiversion           bool                 `json:"multiversion,omitempty"`
	DefaultExperiment      string               `json:"defaultExperiment,omitempty"`
	RunCompletionTTL       *metav1.Duration     `json:"runCompletionTTL,omitempty"`
	RunCompletionFeed      ServiceConfiguration `json:"runCompletionFeed,omitempty"`
}

type Endpoint struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
	Path string `json:"path,omitempty"`
}

func (e Endpoint) URL() string {
	return fmt.Sprintf("%s:%d%s", e.Host, e.Port, e.Path)
}

type ServiceConfiguration struct {
	Port      int        `json:"port,omitempty"`
	Endpoints []Endpoint `json:"endpoints,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type KfpControllerConfig struct {
	metav1.TypeMeta        `json:",inline"`
	metav1.ObjectMeta      `json:"metadata,omitempty"`
	Spec                   KfpControllerConfigSpec `json:"spec,omitempty"`
	apis.ControllerWrapper `json:"controller,omitempty"`
}

// +kubebuilder:object:root=true
func init() {
	SchemeBuilder.Register(&KfpControllerConfig{})
}
