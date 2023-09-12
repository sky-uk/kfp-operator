package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

type Configuration struct {
	PipelineStorage string `json:"pipelineStorage,omitempty"`

	DefaultProvider string `json:"defaultProvider,omitempty"`

	WorkflowTemplatePrefix string `json:"workflowTemplatePrefix,omitempty"`

	WorkflowNamespace string `json:"workflowNamespace,omitempty"`

	Multiversion bool `json:"multiversion,omitempty"`

	DefaultBeamArgs []apis.NamedValue `json:"defaultBeamArgs,omitempty"`

	DefaultExperiment string `json:"defaultExperiment,omitempty"`

	RunCompletionTTL *metav1.Duration `json:"runCompletionTTL,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:storageversion

type KfpControllerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Workflows         Configuration `json:"spec,omitempty"`

	cfg.ControllerManagerConfigurationSpec `json:"controller,omitempty"`
}

//+kubebuilder:object:root=true

func init() {
	SchemeBuilder.Register(&KfpControllerConfig{})
}
