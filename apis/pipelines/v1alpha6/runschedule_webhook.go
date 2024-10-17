package v1alpha6

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

func (rs *RunSchedule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(rs).
		Complete()
}
