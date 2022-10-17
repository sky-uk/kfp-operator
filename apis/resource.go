package apis

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type Resource interface {
	metav1.Object
	runtime.Object
	GetStatus() Status
	SetStatus(Status)
	GetNamespacedName() types.NamespacedName
	ComputeVersion() string
	// GetKind is a workaround to address https://github.com/sky-uk/kfp-operator/issues/137
	GetKind() string
}
