package v1beta1

import (
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (r *Run) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1beta1-run,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=runs,verbs=create;update,versions=v1beta1,name=vrun.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Run{}

func (r *Run) ValidateCreate() (admission.Warnings, error) {
	for i, p := range r.Spec.Parameters {
		if p.ValueFrom != nil && p.Value != "" {
			return nil, apierrors.NewInvalid(r.GroupVersionKind().GroupKind(),
				r.Name, field.ErrorList{
					field.Invalid(
						field.NewPath("spec").Child("parameters").Index(i),
						p,
						"only one of value or valueFrom can be set"),
				})
		}
	}

	return nil, nil
}

func (r *Run) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	if !reflect.DeepEqual(r.Spec, old.(*Run).Spec) {
		return nil, apierrors.NewInvalid(r.GroupVersionKind().GroupKind(),
			r.Name, field.ErrorList{field.Forbidden(field.NewPath("spec"), "immutable")})
	}

	return nil, nil
}

func (r *Run) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
