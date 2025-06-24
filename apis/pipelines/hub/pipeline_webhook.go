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

func NewPipelineValidatorWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&Pipeline{}).
		WithValidator(
			&PipelineValidator{
				reader: mgr.GetClient(),
			},
		).
		Complete()
}

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1beta1-pipeline,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=pipelines,verbs=create;update,versions=v1beta1,name=vpipeline.kb.io,admissionReviewVersions=v1
//+kubebuilder:object:generate=false

type PipelineValidator struct {
	reader client.Reader
}

var _ webhook.CustomValidator = &PipelineValidator{}

func (p *PipelineValidator) ValidateCreate(
	ctx context.Context,
	obj runtime.Object,
) (warnings admission.Warnings, err error) {
	return p.validate(ctx, obj)
}

func (p *PipelineValidator) ValidateUpdate(
	ctx context.Context,
	_, newObj runtime.Object,
) (warnings admission.Warnings, err error) {
	return p.validate(ctx, newObj)
}

func (p *PipelineValidator) ValidateDelete(
	_ context.Context,
	_ runtime.Object,
) (warnings admission.Warnings, err error) {
	return nil, nil
}

func (p *PipelineValidator) validate(ctx context.Context, obj runtime.Object) (warnings admission.Warnings, err error) {
	pipeline, ok := obj.(*Pipeline)

	if !ok {
		return nil, apierrors.NewBadRequest(
			fmt.Sprintf(
				"Got kind=%v; expected kind=%v",
				obj.GetObjectKind().GroupVersionKind().GroupKind(),
				GroupVersion.WithKind(Pipeline{}.GetKind()).GroupKind(),
			),
		)
	}

	provider := Provider{}
	if err = p.reader.Get(
		ctx,
		client.ObjectKey{
			Namespace: pipeline.Spec.Provider.Namespace,
			Name:      pipeline.Spec.Provider.Name,
		},
		&provider,
	); err != nil {
		return nil, apierrors.NewInvalid(
			obj.GetObjectKind().GroupVersionKind().GroupKind(),
			fmt.Sprintf("%s/%s", pipeline.GetNamespacedName().Namespace, pipeline.GetNamespacedName().Name),
			[]*field.Error{
				field.NotFound(
					field.NewPath("spec", "provider"),
					pipeline.Spec.Provider,
				),
			},
		)
	}

	if !lo.Contains(provider.Spec.AllowedNamespaces, pipeline.GetNamespacedName().Namespace) {
		return nil, apierrors.NewInvalid(
			obj.GetObjectKind().GroupVersionKind().GroupKind(),
			fmt.Sprintf("%s/%s", pipeline.GetNamespacedName().Namespace, pipeline.GetNamespacedName().Name),
			[]*field.Error{
				field.Forbidden(
					field.NewPath("metadata", "namespace"),
					fmt.Sprintf("namespace %s is not allowed by provider %s", pipeline.GetNamespacedName().Namespace, provider.GetNamespacedName().String()),
				),
			},
		)
	}

	providerFrameworkNames := lo.Map(provider.Spec.Frameworks, func(f Framework, _ int) string {
		return f.Name
	})

	if !lo.Contains(providerFrameworkNames, pipeline.Spec.Framework.Name) {
		return nil, apierrors.NewInvalid(
			obj.GetObjectKind().GroupVersionKind().GroupKind(),
			fmt.Sprintf("%s/%s", pipeline.GetNamespacedName().Namespace, pipeline.GetNamespacedName().Name),
			[]*field.Error{
				field.NotSupported(
					field.NewPath("spec", "framework"),
					pipeline.Spec.Framework.Name,
					providerFrameworkNames,
				),
			},
		)
	}

	return nil, nil
}
