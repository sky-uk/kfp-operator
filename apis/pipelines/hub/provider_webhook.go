package v1beta1

import (
	"context"
	"fmt"

	"github.com/samber/lo"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

//+kubebuilder:webhook:path=/validate-providers-kubeflow-org-v1beta1-provider,mutating=false,failurePolicy=fail,sideEffects=None,groups=providers.kubeflow.org,resources=providers,verbs=update;delete,versions=v1beta1,name=vprovider.kb.io,admissionReviewVersions=v1
//+kubebuilder:object:generate=false

type ProviderValidator struct {
	reader client.Reader
}

var _ webhook.CustomValidator = &ProviderValidator{}

func NewProviderValidatorWebhook(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(
		context.Background(),
		&Pipeline{},
		".spec.provider",
		func(obj client.Object) []string {
			pipeline, ok := obj.(*Pipeline)
			if !ok {
				return nil
			}
			providerNsnStr, err := pipeline.Spec.Provider.String()
			if err != nil {
				return nil
			}
			return []string{providerNsnStr}
		},
	); err != nil {
		return fmt.Errorf("failed to set up field index on Pipeline.spec.provider: %w", err)
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(&Provider{}).
		WithValidator(
			&ProviderValidator{
				reader: mgr.GetClient(),
			},
		).
		Complete()
}

func (p *ProviderValidator) ValidateCreate(
	ctx context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	return nil, nil
}

func (p *ProviderValidator) ValidateUpdate(
	ctx context.Context,
	_, newObj runtime.Object,
) (admission.Warnings, error) {
	return nil, nil
}

func (p *ProviderValidator) ValidateDelete(
	ctx context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	provider, ok := obj.(*Provider)

	providerPtr := &Provider{}

	if !ok {
		return nil, apierrors.NewBadRequest(
			fmt.Sprintf(
				"Got kind=%v; expected kind=%v",
				obj.GetObjectKind().GroupVersionKind().GroupKind(),
				GroupVersion.WithKind(providerPtr.GetKind()).GroupKind(),
			),
		)

	}

	var pipelineList PipelineList
	// TODOO: implement case for when allowedNamespaces is not there/empty(?)
	for _, ns := range provider.Spec.AllowedNamespaces {
		if err := p.reader.List(
			ctx,
			&pipelineList,
			client.InNamespace(ns),
			client.MatchingFields{
				"spec.provider": provider.GetNamespacedName().String(),
			},
		); err != nil {
			return nil, apierrors.NewInternalError(err)
		}

		pipelinesWithMatchingFramework := lo.Filter(
			pipelineList.Items,
			func(p Pipeline, _ int) bool {
				return lo.Contains(
					lo.Map(
						provider.Spec.Frameworks,
						func(f Framework, _ int) string {
							return f.Name
						},
					),
					p.Spec.Framework.Name,
				)
			},
		)

		var errors field.ErrorList

		for _, pp := range pipelinesWithMatchingFramework {
			errors = append(
				errors,
				field.Forbidden(
					field.NewPath("spec"),
					fmt.Sprintf(
						"no allowed bro due to existing pipeline %v containing framework %v",
						pp.GetNamespacedName(),
						pp.Spec.Framework.Name,
					),
				),
			)
		}

		if len(errors) > 0 {
			return nil, apierrors.NewInvalid(
				provider.GetObjectKind().GroupVersionKind().GroupKind(),
				provider.GetNamespacedName().String(),
				errors,
			)
		}
	}

	return nil, nil
}
