package v1alpha4

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

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1alpha4-run,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=runs,verbs=create;update,versions=v1alpha4,name=vrun.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Run{}

func (r *Run) ValidateCreate() error {
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
