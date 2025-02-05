package predicates

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type ServiceChangedPredicate struct {
	predicate.Funcs
}

// Create Used to filter out create events triggered by a controller operation
func (ServiceChangedPredicate) Create(e event.CreateEvent) bool {
	svc, ok := e.Object.(*corev1.Service)
	if !ok {
		return true
	}

	//Ignore create events caused by controller ops
	if svc.Annotations != nil && svc.Annotations[ControllerManagedKey] == "true" {
		return false
	}

	return true
}
