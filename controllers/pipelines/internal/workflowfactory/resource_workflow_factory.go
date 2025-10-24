package workflowfactory

import (
	"encoding/json"
	"fmt"
	"slices"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/jsonutil"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type WorkflowFactory[R pipelineshub.Resource] interface {
	ConstructCreationWorkflow(
		provider pipelineshub.Provider,
		providerSvc corev1.Service,
		resource R,
	) (*argo.Workflow, error)

	ConstructUpdateWorkflow(
		provider pipelineshub.Provider,
		providerSvc corev1.Service,
		resource R,
	) (*argo.Workflow, error)

	ConstructDeletionWorkflow(
		provider pipelineshub.Provider,
		providerSvc corev1.Service,
		resource R,
	) (*argo.Workflow, error)
}

func createProviderServiceUrl(svc corev1.Service, port int) string {
	return fmt.Sprintf("%s.%s:%d", svc.Name, svc.Namespace, port)
}

type ResourceWorkflowFactory[R pipelineshub.Resource, ResourceDefinition any] struct {
	Config                config.KfpControllerConfigSpec
	TemplateSuffix        string
	DefinitionCreator     func(pipelineshub.Provider, R) ([]pipelineshub.Patch, ResourceDefinition, error)
	WorkflowParamsCreator func(pipelineshub.Provider, R) ([]argo.Parameter, error)
}

const (
	CompiledSuffix = "compiled"
	SimpleSuffix   = "simple"
)

// Template name methods - create and update operations use suffixed templates to differentiate
// between resource types (e.g., "compiled" for pipelines, "simple" for basic resources).
// Delete operations use a single shared template since deletion is generic across all resource types.
func (rwf ResourceWorkflowFactory[R, ResourceDefinition]) createTemplateName() string {
	return fmt.Sprintf("%screate-%s", rwf.Config.WorkflowTemplatePrefix, rwf.TemplateSuffix)
}

func (rwf ResourceWorkflowFactory[R, ResourceDefinition]) updateTemplateName() string {
	return fmt.Sprintf("%supdate-%s", rwf.Config.WorkflowTemplatePrefix, rwf.TemplateSuffix)
}

func (rwf ResourceWorkflowFactory[R, ResourceDefinition]) deleteTemplateName() string {
	return fmt.Sprintf("%sdelete", rwf.Config.WorkflowTemplatePrefix)
}

func WorkflowParamsCreatorNoop[R any](provider pipelineshub.Provider, _ R) ([]argo.Parameter, error) {
	return []argo.Parameter{}, nil
}

func (workflows ResourceWorkflowFactory[R, ResourceDefinition]) CommonWorkflowMeta(
	owner pipelineshub.Resource,
) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", owner.GetKind(), owner.GetName()),
		Namespace:    workflows.Config.WorkflowNamespace,
		Labels:       workflowconstants.CommonWorkflowLabels(owner),
	}
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) resourceDefinitionJson(provider pipelineshub.Provider, resource R) (string, error) {
	patches, resourceDefinition, err := workflows.DefinitionCreator(provider, resource)
	if err != nil {
		return "", err
	}

	marshalled, err := json.Marshal(&resourceDefinition)
	if err != nil {
		return "", err
	}

	patchedJsonString, err := PatchJson(patches, marshalled)
	if err != nil {
		return "", err
	}

	return patchedJsonString, nil
}

func checkResourceNamespaceAllowed(
	resourceNamespacedName types.NamespacedName,
	provider pipelineshub.Provider,
) error {
	if len(provider.Spec.AllowedNamespaces) > 0 && !slices.Contains(provider.Spec.AllowedNamespaces, resourceNamespacedName.Namespace) {
		return fmt.Errorf("resource %s in namespace %s is not allowed by provider %s", resourceNamespacedName.Name, resourceNamespacedName.Namespace, provider.Name)
	}
	return nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructCreationWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.resourceDefinitionJson(provider, resource)
	if err != nil {
		return nil, err
	}

	namespacedProvider, err := provider.GetCommonNamespacedName().String()
	if err != nil {
		return nil, err
	}

	if err = checkResourceNamespaceAllowed(resource.GetNamespacedName(), provider); err != nil {
		return nil, err
	}

	params := []argo.Parameter{
		{
			Name:  workflowconstants.ResourceKindParameterName,
			Value: argo.AnyStringPtr(resource.GetKind()),
		},
		{
			Name:  workflowconstants.ResourceDefinitionParameterName,
			Value: argo.AnyStringPtr(resourceDefinition),
		},
		{
			Name:  workflowconstants.ProviderNameParameterName,
			Value: argo.AnyStringPtr(namespacedProvider),
		},
		{
			Name: workflowconstants.ProviderServiceUrl,
			Value: argo.AnyStringPtr(
				createProviderServiceUrl(
					providerSvc,
					workflows.Config.DefaultProviderValues.ServicePort,
				),
			),
		},
	}

	additionalParams, err := workflows.WorkflowParamsCreator(provider, resource)
	if err != nil {
		return nil, err
	}
	params = append(params, additionalParams...)

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(resource),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: params,
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.createTemplateName(),
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructUpdateWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.resourceDefinitionJson(provider, resource)
	if err != nil {
		return nil, err
	}

	namespacedProvider, err := provider.GetCommonNamespacedName().String()
	if err != nil {
		return nil, err
	}

	if err = checkResourceNamespaceAllowed(resource.GetNamespacedName(), provider); err != nil {
		return nil, err
	}

	params := []argo.Parameter{
		{
			Name:  workflowconstants.ResourceKindParameterName,
			Value: argo.AnyStringPtr(resource.GetKind()),
		},
		{
			Name:  workflowconstants.ResourceDefinitionParameterName,
			Value: argo.AnyStringPtr(resourceDefinition),
		},
		{
			Name:  workflowconstants.ResourceIdParameterName,
			Value: argo.AnyStringPtr(resource.GetStatus().Provider.Id),
		},
		{
			Name:  workflowconstants.ProviderNameParameterName,
			Value: argo.AnyStringPtr(namespacedProvider),
		},
		{
			Name: workflowconstants.ProviderServiceUrl,
			Value: argo.AnyStringPtr(
				createProviderServiceUrl(
					providerSvc,
					workflows.Config.DefaultProviderValues.ServicePort,
				),
			),
		},
	}

	additionalParams, err := workflows.WorkflowParamsCreator(provider, resource)
	if err != nil {
		return nil, err
	}
	params = append(params, additionalParams...)

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(resource),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: params,
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.updateTemplateName(),
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructDeletionWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	namespacedProvider, err := provider.GetCommonNamespacedName().String()
	if err != nil {
		fmt.Println("ResourceWorkflowFactory: err: ", err)
		return nil, err
	}

	if err = checkResourceNamespaceAllowed(resource.GetNamespacedName(), provider); err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(resource),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  workflowconstants.ResourceKindParameterName,
						Value: argo.AnyStringPtr(resource.GetKind()),
					},
					{
						Name:  workflowconstants.ResourceIdParameterName,
						Value: argo.AnyStringPtr(resource.GetStatus().Provider.Id),
					},
					{
						Name:  workflowconstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(namespacedProvider),
					},
					{
						Name: workflowconstants.ProviderServiceUrl,
						Value: argo.AnyStringPtr(
							createProviderServiceUrl(
								providerSvc,
								workflows.Config.DefaultProviderValues.ServicePort,
							),
						),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.deleteTemplateName(),
			},
		},
	}, nil
}
