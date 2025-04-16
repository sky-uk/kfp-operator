package v1beta1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (p *Provider) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(p).
		Complete()
}

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1beta1-provider,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=providers,verbs=create;update,versions=v1beta1,name=provider.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Provider{}

func (p *Provider) validateUniqueStructures() (errors field.ErrorList) {
	return
}

func (p *Provider) validateRuntimeParameters() (errors field.ErrorList) {
	return
}

func (p *Provider) validate() (admission.Warnings, error) {
	return nil, nil
}

func (p *Provider) ValidateCreate() (admission.Warnings, error) {
	return p.validate()
}

func (p *Provider) ValidateUpdate(_ runtime.Object) (admission.Warnings, error) {
	return p.validate()
}

func (p *Provider) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}
