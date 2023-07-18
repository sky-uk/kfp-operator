package v1alpha5

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

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1alpha5-runconfiguration,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=runconfigurations,verbs=create;update,versions=v1alpha5,name=vrunconfiguration.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &RunConfiguration{}

func (rc *RunConfiguration) validateUniqueStructures() error {
	uniqueScheduled := pipelines.Unique(rc.Spec.Triggers.Schedules)
	if len(rc.Spec.Triggers.Schedules) > len(uniqueScheduled) {
		return apierrors.NewInvalid(rc.GroupVersionKind().GroupKind(),
			rc.Name, field.ErrorList{field.Duplicate(field.NewPath("spec").Key("triggers").Key("schedules"), rc.Spec.Triggers.Schedules)})
	}

	uniqueOnChangeTriggers := pipelines.Unique(rc.Spec.Triggers.OnChange)
	if len(rc.Spec.Triggers.OnChange) > len(uniqueOnChangeTriggers) {
		return apierrors.NewInvalid(rc.GroupVersionKind().GroupKind(),
			rc.Name, field.ErrorList{field.Duplicate(field.NewPath("spec").Key("triggers").Key("onChange"), rc.Spec.Triggers.OnChange)})
	}

	return nil
}

func (rc *RunConfiguration) validateRuntimeParameters() error {
	for i, rp := range rc.Spec.Run.RuntimeParameters {
		if rp.ValueFrom != nil && rp.Value != "" {
			return apierrors.NewInvalid(rc.GroupVersionKind().GroupKind(),
				rc.Name, field.ErrorList{
					field.Invalid(
						field.NewPath("spec").Child("run").Child("runtimeParameters").Index(i),
						rp,
						"only one of value or valueFrom can be set"),
				})
		}
	}

	return nil
}

func (rc *RunConfiguration) validate() error {
	if err := rc.validateRuntimeParameters(); err != nil {
		return err
	}

	return rc.validateUniqueStructures()
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
