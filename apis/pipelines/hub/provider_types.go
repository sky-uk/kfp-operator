package v1beta1

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/argo/common"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName="mlprv"
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="SynchronizationState",type="string",JSONPath=".status.conditions[?(@.type == 'Synchronized')].reason"
// +kubebuilder:storageversion
// +kubebuilder:pruning:PreserveUnknownFields
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProviderSpec `json:"spec,omitempty"`
	Status Status       `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

type ProviderSpec struct {
	ServiceImage string `json:"serviceImage" yaml:"serviceImage"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:deprecatedversion=true
	// +kubebuilder:default=""
	// Deprecated: This field is ignored and will be removed in future versions.
	ExecutionMode string `json:"executionMode" yaml:"executionMode"`
	// +kubebuilder:validation:Pattern:=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	ServiceAccount      string                           `json:"serviceAccount" yaml:"serviceAccount"`
	PipelineRootStorage string                           `json:"pipelineRootStorage" yaml:"pipelineRootStorage"`
	Parameters          map[string]*apiextensionsv1.JSON `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Frameworks          []Framework                      `json:"frameworks,omitempty" yaml:"frameworks,omitempty"`
	AllowedNamespaces   []string                         `json:"allowedNamespaces,omitempty" yaml:"allowedNamespaces,omitempty"`
}

type Framework struct {
	Name    string  `json:"name,omitempty" yaml:"name,omitempty"`
	Image   string  `json:"image,omitempty" yaml:"image,omitempty"`
	Patches []Patch `json:"patches,omitempty" yaml:"patches,omitempty"`
}

type Patch struct {
	// +kubebuilder:validation:Enum=json;merge
	Type    string `json:"type,omitempty" yaml:"type,omitempty"`
	Payload string `json:"payload,omitempty" yaml:"payload,omitempty"`
}

func (p *Provider) ComputeVersion() string {
	// Not used by Provider controller but required to satisfy Resource interface
	return ""
}

func (p *Provider) GetStatus() Status {
	return p.Status
}

func (p *Provider) SetStatus(status Status) {
	p.Status = status
}

func (p *Provider) GetNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      p.Name,
		Namespace: p.Namespace,
	}
}

func (p Provider) GetCommonNamespacedName() common.NamespacedName {
	return common.NamespacedName{
		Name:      p.Name,
		Namespace: p.Namespace,
	}
}

func (p *Provider) GetKind() string {
	return "provider"
}

func (p *Provider) StatusWithCondition(state apis.SynchronizationState, message string) {
	p.Status.Conditions = p.Status.Conditions.MergeIntoConditions(metav1.Condition{
		LastTransitionTime: metav1.Now().Rfc3339Copy(),
		Message:            message,
		ObservedGeneration: p.Status.ObservedGeneration,
		Type:               apis.ConditionTypes.SynchronizationSucceeded,
		Status:             apis.ConditionStatusForSynchronizationState(state),
		Reason:             string(state),
	})
}

func init() {
	SchemeBuilder.Register(&Provider{}, &ProviderList{})
}
