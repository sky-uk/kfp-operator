package v2

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"
)

type Configuration struct {
	PipelineStorage string `json:"pipelineStorage,omitempty"`
	KfpEndpoint     string `json:"kfpEndpoint,omitempty"`

	Argo ArgoConfiguration `json:"argo,omitempty"`

	DefaultBeamArgs map[string]string `json:"defaultBeamArgs,omitempty"`

	DefaultExperiment string `json:"defaultExperiment,omitempty"`

	Debug pipelinesv1.DebugOptions `json:"debug,omitempty"`
}

type ArgoConfiguration struct {
	CompilerImage  string `json:"compilerImage,omitempty"`
	KfpSdkImage    string `json:"kfpSdkImage,omitempty"`
	ServiceAccount string `json:"serviceAccount,omitempty"`

	ContainerDefaults apiv1.Container `json:"containerDefaults,omitempty"`
	MetadataDefaults  argo.Metadata   `json:"metadataDefaults,omitempty"`
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
