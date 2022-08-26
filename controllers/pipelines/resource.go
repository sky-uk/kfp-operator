package pipelines

import (
	"github.com/sky-uk/kfp-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type Resource interface {
	metav1.Object
	runtime.Object
	GetStatus() apis.Status
	SetStatus(apis.Status)
	GetNamespacedName() types.NamespacedName
	// GetKind is a workaround to address https://github.com/sky-uk/kfp-operator/issues/137
	GetKind() string
}
