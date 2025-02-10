package predicates

import (
	v1 "k8s.io/kubernetes/staging/src/k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type DeploymentChangedPredicate struct {
	predicate.Funcs
}

func (DeploymentChangedPredicate) Create(e event.CreateEvent) bool {
	svc, ok := e.Object.(*v1.Deployment)
	if !ok {
		return true
	}

	//Ignore create events caused by controller ops
	if svc.Annotations != nil && svc.Annotations[ControllerManagedKey] == "true" {
		return false
	}

	return true
}
