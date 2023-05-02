package v1alpha5

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// +kubebuilder:object:generate=false
type Resource interface {
	HasProvider
	runtime.Object
	GetStatus() Status
	SetStatus(Status)
	GetNamespacedName() types.NamespacedName
	ComputeVersion() string
	// GetKind is a workaround to address https://github.com/sky-uk/kfp-operator/issues/137
	GetKind() string
}

// +kubebuilder:object:generate=false
type HasProvider interface {
	metav1.Object
	GetProvider() string
}
