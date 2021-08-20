package external

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
)

func InitSchemes(scheme *runtime.Scheme) error {
	return argo.AddToScheme(scheme)
}
