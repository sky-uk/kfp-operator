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
				client: mgr.GetClient(),
			},
		).
		Complete()
}

//+kubebuilder:webhook:path=/validate-pipelines-kubeflow-org-v1beta1-pipeline,mutating=false,failurePolicy=fail,sideEffects=None,groups=pipelines.kubeflow.org,resources=pipelines,verbs=create;update,versions=v1beta1,name=vpipeline.kb.io,admissionReviewVersions=v1
//+kubebuilder:object:generate=false

type PipelineValidator struct {
	client client.Client
}

var _ webhook.CustomValidator = &PipelineValidator{}

func (p *PipelineValidator) ValidateCreate(
	ctx context.Context,
	obj runtime.Object,
) (warnings admission.Warnings, err error) {
	pipeline, ok := obj.(*Pipeline)

	if !ok {
		return nil, apierrors.NewInvalid(
			obj.GetObjectKind().GroupVersionKind().GroupKind(),
			"dunno",
			[]*field.Error{
				field.TypeInvalid(
					field.NewPath("kind"),
					obj,
					"incorrect kind",
				),
			},
		)
	}

	provider := Provider{}
	if err = p.client.Get(
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

func (p *PipelineValidator) ValidateUpdate(
	ctx context.Context,
	oldObj, newObj runtime.Object,
) (warnings admission.Warnings, err error) {
	return nil, nil
}

func (p *PipelineValidator) ValidateDelete(
	ctx context.Context,
	obj runtime.Object,
) (warnings admission.Warnings, err error) {
	return nil, nil
}
