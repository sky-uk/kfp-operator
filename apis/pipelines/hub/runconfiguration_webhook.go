package v1beta1

import (
	"github.com/sky-uk/kfp-operator/apis"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (rc *RunConfiguration) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(rc).
		Complete()
}

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1beta1-runconfiguration,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=create;update,versions=v1beta1,name=vrunconfiguration.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &RunConfiguration{}

func (rc *RunConfiguration) validateUniqueStructures() (errors field.ErrorList) {
	duplicateSchedules := apis.Duplicates(rc.Spec.Triggers.Schedules)
	schedulePath := field.NewPath("spec").Key("triggers").Key("schedules")
	errors = append(errors, apis.Map(duplicateSchedules, func(schedule Schedule) *field.Error {
		return field.Duplicate(schedulePath, schedule)
	})...)

	duplicateOnChangeTriggers := apis.Duplicates(rc.Spec.Triggers.OnChange)
	onChangePath := field.NewPath("spec").Key("triggers").Key("onChange")
	errors = append(errors, apis.Map(duplicateOnChangeTriggers, func(onChange OnChangeType) *field.Error {
		return field.Duplicate(onChangePath, onChange)
	})...)

	return
}

func (rc *RunConfiguration) validateRuntimeParameters() (errors field.ErrorList) {
	runtimeParametersPath := field.NewPath("spec").Child("run").Child("runtimeParameters")
	for i, rp := range rc.Spec.Run.RuntimeParameters {
		if rp.ValueFrom != nil && rp.Value != "" {
			errors = append(errors,
				field.Invalid(runtimeParametersPath.Index(i), rp, "only one of value or valueFrom can be set"),
			)
		}
	}

	return
}

func (rc *RunConfiguration) validate() (admission.Warnings, error) {
	errors := apis.Flatten(rc.validateRuntimeParameters(), rc.validateUniqueStructures())

	if len(errors) > 0 {
		return nil, apierrors.NewInvalid(rc.GroupVersionKind().GroupKind(), rc.Name, errors)
	}

	return nil, nil
}

func (rc *RunConfiguration) ValidateCreate() (admission.Warnings, error) {
	return rc.validate()
}

func (rc *RunConfiguration) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	return rc.validate()
}

func (rc *RunConfiguration) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
