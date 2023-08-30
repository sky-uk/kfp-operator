package v1alpha5

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *Run) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

var _ webhook.Validator = &Run{}

func (r *Run) ValidateCreate() error {
	for i, rp := range r.Spec.RuntimeParameters {
		if rp.ValueFrom != nil && rp.Value != "" {
			return apierrors.NewInvalid(r.GroupVersionKind().GroupKind(),
				r.Name, field.ErrorList{
					field.Invalid(
						field.NewPath("spec").Child("runtimeParameters").Index(i),
						rp,
						"only one of value or valueFrom can be set"),
				})
		}
	}

	return nil
}

func (r *Run) ValidateUpdate(old runtime.Object) error {
	if !reflect.DeepEqual(r.Spec, old.(*Run).Spec) {
		return apierrors.NewInvalid(r.GroupVersionKind().GroupKind(),
			r.Name, field.ErrorList{field.Forbidden(field.NewPath("spec"), "immutable")})
	}

	return nil
}

func (r *Run) ValidateDelete() error {
	return nil
}
