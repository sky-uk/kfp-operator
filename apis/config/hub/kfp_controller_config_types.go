package v1beta1

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:storageversion
type KfpControllerConfig struct {
	metav1.TypeMeta        `json:",inline"`
	metav1.ObjectMeta      `json:"metadata,omitempty"`
	Spec                   KfpControllerConfigSpec `json:"spec,omitempty"`
	apis.ControllerWrapper `json:"controller,omitempty"`
}

type KfpControllerConfigSpec struct {
	DefaultProvider        string                `json:"defaultProvider,omitempty"`
	DefaultProviderValues  DefaultProviderValues `json:"defaultProviderValues,omitempty"`
	DefaultTfxImage        string                `json:"defaultTfxImage,omitempty"`
	WorkflowTemplatePrefix string                `json:"workflowTemplatePrefix,omitempty"`
	WorkflowNamespace      string                `json:"workflowNamespace,omitempty"`
	Multiversion           bool                  `json:"multiversion,omitempty"`
	DefaultExperiment      string                `json:"defaultExperiment,omitempty"`
	RunCompletionTTL       *metav1.Duration      `json:"runCompletionTTL,omitempty"`
	RunCompletionFeed      ServiceConfiguration  `json:"runCompletionFeed,omitempty"`
}

type DefaultProviderValues struct {
	Labels               map[string]string  `json:"labels,omitempty"`
	Replicas             int                `json:"replicas,omitempty"`
	PodTemplateSpec      v1.PodTemplateSpec `json:"podTemplateSpec,omitempty"`
	ServiceContainerName string             `json:"serviceContainerName,omitempty"`
	ServicePort          int                `json:"servicePort,omitempty"`
	MetricsPort          int                `json:"metricsPort,omitempty"`
}

type ServiceConfiguration struct {
	Port      int        `json:"port,omitempty"`
	Endpoints []Endpoint `json:"endpoints,omitempty"`
}

type Endpoint struct {
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
	Path string `json:"path,omitempty"`
}

func (e Endpoint) URL() string {
	return fmt.Sprintf("%s:%d%s", e.Host, e.Port, e.Path)
}

// +kubebuilder:object:root=true
func init() {
	SchemeBuilder.Register(&KfpControllerConfig{})
}
