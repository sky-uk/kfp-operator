package v1alpha5

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

func (rs *RunSchedule) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(rs).
		Complete()
}
