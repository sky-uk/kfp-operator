package v1alpha5

import (
	"fmt"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var runconfigurationlog = logf.Log.WithName("runconfiguration-resource")

func (r *RunConfiguration) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1alpha5-runconfiguration,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=create;update,versions=v1alpha5,name=vrunconfiguration.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &RunConfiguration{}

func (rc *RunConfiguration) ValidateCreate() error {
	return rc.validateTriggers()
}

func (rc *RunConfiguration) ValidateUpdate(_ runtime.Object) error {
	return rc.validateTriggers()
}

func (rc *RunConfiguration) ValidateDelete() error {
	return nil
}

func (rc *RunConfiguration) validateTriggers() error {
	if errors := rc.errorsInTriggers(); errors != nil {
		return apierrors.NewInvalid(rc.GroupVersionKind().GroupKind(), rc.Name, errors)
	}
	return nil
}

func (rc *RunConfiguration) errorsInTriggers() field.ErrorList {
	var errors field.ErrorList
	for i, trigger := range rc.Spec.Triggers {
		switch trigger.Type {
		case TriggerTypes.Schedule:
			if trigger.CronExpression == "" {
				detail := fmt.Sprintf("required for trigger type %s", trigger.Type)
				errors = append(errors, field.Required(
					triggerFieldPath(i).Child("cronExpression"),
					detail,
				))
			}
		case TriggerTypes.OnChange:
			if trigger.CronExpression != "" {
				detail := fmt.Sprintf("not allowed for trigger type %s", trigger.Type)
				errors = append(errors, field.Forbidden(
					triggerFieldPath(i).Child("cronExpression"),
					detail,
				))
			}
		default:
			errors = append(errors, field.NotSupported(
				triggerFieldPath(i).Child("type"), trigger.Type, nil,
			))
		}
	}
	return errors
}

func triggerFieldPath(i int) *field.Path {
	return field.NewPath("spec").Child("triggers").Index(i)
}
