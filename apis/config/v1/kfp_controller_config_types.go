package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

type Configuration struct {
	PipelineStorage string `json:"pipelineStorage,omitempty"`
	KfpEndpoint     string `json:"kfpEndpoint,omitempty"`
	ServiceAccount  string `json:"serviceAccount,omitempty"`

	CompilerImage   string `json:"compilerImage,omitempty"`
	KfpToolsImage   string `json:"kfpToolsImage,omitempty"`
	ImagePullPolicy string `json:"imagePullPolicy,omitempty"`

	DefaultBeamArgs map[string]string `json:"defaultBeamArgs,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

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
