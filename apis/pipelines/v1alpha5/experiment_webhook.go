package v1alpha5

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

func (e *Experiment) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(e).
		Complete()
}
