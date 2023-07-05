package v1alpha5

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (rc *RunConfiguration) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(rc).
		Complete()
}

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1alpha5-runconfiguration,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=create;update,versions=v1alpha5,name=vrunconfiguration.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &RunConfiguration{}

func (r *RunConfiguration) validate() error {
	for i, rp := range r.Spec.Run.RuntimeParameters {
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

func (r *RunConfiguration) ValidateCreate() error {
	return r.validate()
}

func (r *RunConfiguration) ValidateUpdate(_ runtime.Object) error {
	return r.validate()
}

func (r *RunConfiguration) ValidateDelete() error {
	return nil
}
