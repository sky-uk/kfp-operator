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
	oldObj, newObj runtime.Object,
) (admission.Warnings, error) {
	oldProvider, err := p.castToProviderOrErr(oldObj)
	if err != nil {
		return nil, err
	}

	newProvider, err := p.castToProviderOrErr(newObj)
	if err != nil {
		return nil, err
	}

	deletedFrameworks := lo.Filter(oldProvider.Frameworks(), func(framework string, _ int) bool {
		return !lo.Contains(newProvider.Frameworks(), framework)
	})

	if len(deletedFrameworks) > 0 {
		var pipelineList PipelineList
		if err := p.reader.List(
			ctx,
			&pipelineList,
			client.MatchingFields{
				"spec.provider": newProvider.GetNamespacedName().String(),
			},
		); err != nil {
			return nil, apierrors.NewInternalError(err)
		}

		pipelinesFilteredByNamespaces := filterByNamespaces(pipelineList.Items, newProvider.Spec.AllowedNamespaces)
		pipelinesWithMatchingDeletedFramework := filterByFrameworks(pipelinesFilteredByNamespaces, deletedFrameworks)
		errors := createErrorsForPipelines(pipelinesWithMatchingDeletedFramework, newProvider, "update")

		if len(errors) > 0 {
			return nil, apierrors.NewInvalid(
				newProvider.GetObjectKind().GroupVersionKind().GroupKind(),
				newProvider.GetNamespacedName().String(),
				errors,
			)
		}
	}

	return nil, nil
}

func (p *ProviderValidator) ValidateDelete(
	ctx context.Context,
	obj runtime.Object,
) (admission.Warnings, error) {
	provider, err := p.castToProviderOrErr(obj)
	if err != nil {
		return nil, err
	}

	var pipelineList PipelineList
	if err := p.reader.List(
		ctx,
		&pipelineList,
		client.MatchingFields{
			"spec.provider": provider.GetNamespacedName().String(),
		},
	); err != nil {
		return nil, apierrors.NewInternalError(err)
	}

	pipelinesFilteredByNamespaces := filterByNamespaces(pipelineList.Items, provider.Spec.AllowedNamespaces)
	pipelinesWithMatchingFramework := filterByFrameworks(pipelinesFilteredByNamespaces, provider.Frameworks())
	errors := createErrorsForPipelines(pipelinesWithMatchingFramework, provider, "delete")

	if len(errors) > 0 {
		return nil, apierrors.NewInvalid(
			provider.GetObjectKind().GroupVersionKind().GroupKind(),
			provider.GetNamespacedName().String(),
			errors,
		)
	}

	return nil, nil
}

func (p *ProviderValidator) castToProviderOrErr(obj runtime.Object) (*Provider, error) {
	providerPtr := &Provider{}
	provider, ok := obj.(*Provider)
	if !ok {
		return nil, apierrors.NewBadRequest(
			fmt.Sprintf(
				"Got kind=%v; expected kind=%v",
				obj.GetObjectKind().GroupVersionKind().GroupKind(),
				GroupVersion.WithKind(providerPtr.GetKind()).GroupKind(),
			),
		)
	}
	return provider, nil
}

func createErrorsForPipelines(pipelinesWithMatchingFramework []Pipeline, provider *Provider, operation string) field.ErrorList {
	errors := field.ErrorList{}
	for _, pipeline := range pipelinesWithMatchingFramework {
		errors = append(
			errors,
			field.Forbidden(
				field.NewPath("spec"),
				fmt.Sprintf(
					"Cannot %s provider %v due to existing pipeline %v containing framework %v",
					operation,
					provider.GetNamespacedName(),
					pipeline.GetNamespacedName(),
					pipeline.Spec.Framework.Name,
				),
			),
		)
	}
	return errors
}

func filterByNamespaces(pipelineListItems []Pipeline, namespaces []string) []Pipeline {
	if len(namespaces) > 0 {
		pipelineListItems = lo.Filter(
			pipelineListItems,
			func(pipeline Pipeline, _ int) bool {
				return lo.Contains(
					namespaces,
					pipeline.Namespace,
				)
			},
		)
	}
	return pipelineListItems
}

func filterByFrameworks(pipelineListItems []Pipeline, frameworks []string) []Pipeline {
	pipelinesWithMatchingFramework := lo.Filter(
		pipelineListItems,
		func(pipeline Pipeline, _ int) bool {
			return lo.Contains(
				frameworks,
				pipeline.Spec.Framework.Name,
			)
		},
	)
	return pipelinesWithMatchingFramework
}
