package v1alpha5

import (
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

type Configuration struct {
	DefaultProvider string `json:"defaultProvider,omitempty"`

	WorkflowTemplatePrefix string `json:"workflowTemplatePrefix,omitempty"`

	WorkflowNamespace string `json:"workflowNamespace,omitempty"`

	Multiversion bool `json:"multiversion,omitempty"`

	DefaultExperiment string `json:"defaultExperiment,omitempty"`

	RunCompletionTTL *metav1.Duration `json:"runCompletionTTL,omitempty"`
}

type Endpoint struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
	Path string `json:"path,omitempty"`
}

func (e Endpoint) URL() string {
	return fmt.Sprintf("http://%s:%d/%s", e.Host, e.Port, e.Path)
}

type ServiceConfiguration struct {
	Host      string     `json:"host,omitempty"`
	Port      int        `json:"port,omitempty"`
	Endpoints []Endpoint `json:"endpoints,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

type KfpControllerConfig struct {
	metav1.TypeMeta                        `json:",inline"`
	metav1.ObjectMeta                      `json:"metadata,omitempty"`
	Workflows                              Configuration        `json:"spec,omitempty"`
	RunCompletionFeed                      ServiceConfiguration `json:"runCompletionFeed,omitempty"`
	cfg.ControllerManagerConfigurationSpec `json:"controller,omitempty"`
}

//+kubebuilder:object:root=true

func init() {
	SchemeBuilder.Register(&KfpControllerConfig{})
}
