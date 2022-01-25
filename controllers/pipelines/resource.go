package pipelines

import (
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type Resource interface {
	metav1.Object
	runtime.Object
	GetStatus() pipelinesv1.Status
	SetStatus(pipelinesv1.Status)
}
