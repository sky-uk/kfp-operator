package v1beta1

import (
	"context"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func NewRunValidatorWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &Run{}).
		WithValidator(&RunValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1beta1-run,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=runs,verbs=create;update,versions=v1beta1,name=vrun.kb.io,admissionReviewVersions=v1
// +kubebuilder:object:generate=false

type RunValidator struct{}

func (*RunValidator) ValidateCreate(
	ctx context.Context,
	r *Run,
) (admission.Warnings, error) {
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

func (*RunValidator) ValidateUpdate(
	ctx context.Context,
	oldRun *Run,
	newRun *Run,
) (admission.Warnings, error) {
	if !reflect.DeepEqual(newRun.Spec, oldRun.Spec) {
		return nil, apierrors.NewInvalid(newRun.GroupVersionKind().GroupKind(),
			newRun.Name, field.ErrorList{field.Forbidden(field.NewPath("spec"), "immutable")})
	}
	return nil, nil
}

func (*RunValidator) ValidateDelete(
	_ context.Context,
	_ *Run,
) (admission.Warnings, error) {
	return nil, nil
}
