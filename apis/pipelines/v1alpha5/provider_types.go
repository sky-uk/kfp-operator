package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:object:root=true
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="mlprv"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.conditions[?(@.type == 'Synchronized')].reason"
// +kubebuilder:pruning:PreserveUnknownFields
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderSpec `json:"spec,omitempty"`
	Status Status       `json:"status,omitempty"`
}

type ProviderSpec struct {
	Image         string `json:"image" yaml:"image"`
	ExecutionMode string `json:"executionMode" yaml:"executionMode"`
	// +kubebuilder:validation:Pattern:=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	ServiceAccount      string                           `json:"serviceAccount" yaml:"serviceAccount"`
	DefaultBeamArgs     []apis.NamedValue                `json:"defaultBeamArgs,omitempty" yaml:"defaultBeamArgs,omitempty"`
	PipelineRootStorage string                           `json:"pipelineRootStorage" yaml:"pipelineRootStorage"`
	Parameters          map[string]*apiextensionsv1.JSON `json:"parameters,omitempty" yaml:"parameters,omitempty"`
}

func (p Provider) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      p.Name,
		Namespace: p.Namespace,
	}
}

func (e Provider) GetKind() string {
	return "provider"
}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}
