package v2

import (
	"github.com/sky-uk/kfp-operator/controllers/pipelines/pipeline_workflows"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

type KfpControllerConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Workflows         pipeline_workflows.Configuration `json:"workflows,omitempty"`

	cfg.ControllerManagerConfigurationSpec `json:",inline"`
}

//+kubebuilder:object:root=true

func init() {
	SchemeBuilder.Register(&KfpControllerConfig{})
}
