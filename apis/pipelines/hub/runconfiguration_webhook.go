package v1beta1

import (
	"context"

	"github.com/samber/lo"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func NewRunConfigurationValidatorWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &RunConfiguration{}).
		WithValidator(&RunConfigurationValidator{}).
		Complete()
}

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1beta1-runconfiguration,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=create;update,versions=v1beta1,name=vrunconfiguration.kb.io,admissionReviewVersions=v1
// +kubebuilder:object:generate=false

type RunConfigurationValidator struct{}

func (rc *RunConfiguration) validateUniqueStructures() (errors field.ErrorList) {
	duplicateSchedules := lo.FindDuplicates(rc.Spec.Triggers.Schedules)
	schedulePath := field.NewPath("spec").Key("triggers").Key("schedules")
	errors = append(errors, lo.Map(duplicateSchedules, func(schedule Schedule, _ int) *field.Error {
		return field.Duplicate(schedulePath, schedule)
	})...)

	duplicateOnChangeTriggers := lo.FindDuplicates(rc.Spec.Triggers.OnChange)
	onChangePath := field.NewPath("spec").Key("triggers").Key("onChange")
	errors = append(errors, lo.Map(duplicateOnChangeTriggers, func(onChange OnChangeType, _ int) *field.Error {
		return field.Duplicate(onChangePath, onChange)
	})...)

	return
}

func (rc *RunConfiguration) validateRunParameters() (errors field.ErrorList) {
	runParametersPath := field.NewPath("spec").Child("run").Child("parameters")
	for i, p := range rc.Spec.Run.Parameters {
		if p.ValueFrom != nil && p.Value != "" {
			errors = append(errors,
				field.Invalid(runParametersPath.Index(i), p, "only one of value or valueFrom can be set"),
			)
		}
	}

	return
}

func (rc *RunConfiguration) validate() (admission.Warnings, error) {
	errors := append(rc.validateRunParameters(), rc.validateUniqueStructures()...)

	if len(errors) > 0 {
		return nil, apierrors.NewInvalid(rc.GroupVersionKind().GroupKind(), rc.Name, errors)
	}

	return nil, nil
}

func (*RunConfigurationValidator) ValidateCreate(
	ctx context.Context,
	rc *RunConfiguration,
) (admission.Warnings, error) {
	return rc.validate()
}

func (*RunConfigurationValidator) ValidateUpdate(
	ctx context.Context,
	_ *RunConfiguration,
	rc *RunConfiguration,
) (admission.Warnings, error) {
	return rc.validate()
}

func (*RunConfigurationValidator) ValidateDelete(
	_ context.Context,
	_ *RunConfiguration,
) (admission.Warnings, error) {
	return nil, nil
}
