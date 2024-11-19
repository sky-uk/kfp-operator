package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
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

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1alpha6-runconfiguration,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=create;update,versions=v1alpha6,name=vrunconfiguration.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &RunConfiguration{}

func (rc *RunConfiguration) validateUniqueStructures() (errors field.ErrorList) {
	duplicateSchedules := pipelines.Duplicates(rc.Spec.Triggers.Schedules)
	schedulePath := field.NewPath("spec").Key("triggers").Key("schedules")
	errors = append(errors, pipelines.Map(duplicateSchedules, func(schedule Schedule) *field.Error {
		return field.Duplicate(schedulePath, schedule)
	})...)

	duplicateOnChangeTriggers := pipelines.Duplicates(rc.Spec.Triggers.OnChange)
	onChangePath := field.NewPath("spec").Key("triggers").Key("onChange")
	errors = append(errors, pipelines.Map(duplicateOnChangeTriggers, func(onChange OnChangeType) *field.Error {
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

func (rc *RunConfiguration) validate() error {
	errors := pipelines.Flatten(rc.validateRuntimeParameters(), rc.validateUniqueStructures())

	if len(errors) > 0 {
		return apierrors.NewInvalid(rc.GroupVersionKind().GroupKind(), rc.Name, errors)
	}

	return nil
}

func (rc *RunConfiguration) ValidateCreate() error {
	return rc.validate()
}

func (rc *RunConfiguration) ValidateUpdate(_ runtime.Object) error {
	return rc.validate()
}

func (rc *RunConfiguration) ValidateDelete() error {
	return nil
}
