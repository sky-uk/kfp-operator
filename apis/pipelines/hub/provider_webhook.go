package v1beta1

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (p *Provider) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, p).
		Complete()
}

// Reserved env var names derived by the operator in
// controllers/pipelines.populateServiceContainer. Keep this list in sync with
// the ProviderNameEnvVar/PipelineRootStorageEnvVar constants and the
// PARAMETERS_ prefix defined there; they cannot be imported here without an
// import cycle (controllers depend on this API package).
const (
	reservedProviderNameEnvVar        = "PROVIDERNAME"
	reservedPipelineRootStorageEnvVar = "PIPELINEROOTSTORAGE"
	reservedParametersPrefix          = "PARAMETERS_"
)

func NewProviderValidatorWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &Provider{}).
		WithValidator(&ProviderValidator{}).
		Complete()
}

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1beta1-provider,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=providers,verbs=create;update,versions=v1beta1,name=vprovider.kb.io,admissionReviewVersions=v1
//+kubebuilder:object:generate=false

type ProviderValidator struct{}

func (v *ProviderValidator) ValidateCreate(
	_ context.Context,
	provider *Provider,
) (admission.Warnings, error) {
	return v.validate(provider)
}

func (v *ProviderValidator) ValidateUpdate(
	_ context.Context,
	_ *Provider,
	newObj *Provider,
) (admission.Warnings, error) {
	return v.validate(newObj)
}

func (v *ProviderValidator) ValidateDelete(
	_ context.Context,
	_ *Provider,
) (admission.Warnings, error) {
	return nil, nil
}

func isReservedEnvVar(name string) bool {
	upper := strings.ToUpper(name)
	return upper == reservedProviderNameEnvVar ||
		upper == reservedPipelineRootStorageEnvVar ||
		strings.HasPrefix(upper, reservedParametersPrefix)
}

func (v *ProviderValidator) validate(
	provider *Provider,
) (admission.Warnings, error) {
	var errs field.ErrorList
	for i, envVar := range provider.Spec.PodTemplateEnv {
		if isReservedEnvVar(envVar.Name) {
			errs = append(errs, field.Forbidden(
				field.NewPath("spec", "podTemplateEnv").Index(i).Child("name"),
				fmt.Sprintf(
					"environment variable %q is reserved by the operator and cannot be set via podTemplateEnv",
					envVar.Name,
				),
			))
		}
	}

	if len(errs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(
		provider.GetObjectKind().GroupVersionKind().GroupKind(),
		provider.GetNamespacedName().String(),
		errs,
	)
}
