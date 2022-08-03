package pipelines

import (
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type Resource interface {
	metav1.Object
	runtime.Object
	GetStatus() pipelinesv1.Status
	SetStatus(pipelinesv1.Status)
	GetNamespacedName() types.NamespacedName
	// GetKind is a workaround to address https://github.com/sky-uk/kfp-operator/issues/137
	GetKind() string
}
