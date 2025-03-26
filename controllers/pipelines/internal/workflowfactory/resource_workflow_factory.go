package workflowfactory

import (
	"encoding/json"
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

type TemplateNameGenerator interface {
	CreateTemplate() string
	UpdateTemplate() string
	DeleteTemplate() string
}

type SuffixedTemplateNameGenerator struct {
	config config.KfpControllerConfigSpec
	suffix string
}

func CompiledTemplateNameGenerator(config config.KfpControllerConfigSpec) TemplateNameGenerator {
	return SuffixedTemplateNameGenerator{config: config, suffix: "compiled"}
}

func SimpleTemplateNameGenerator(config config.KfpControllerConfigSpec) TemplateNameGenerator {
	return SuffixedTemplateNameGenerator{config: config, suffix: "simple"}
}

func (stng SuffixedTemplateNameGenerator) CreateTemplate() string {
	return fmt.Sprintf("%screate-%s", stng.config.WorkflowTemplatePrefix, stng.suffix)
}

func (stng SuffixedTemplateNameGenerator) UpdateTemplate() string {
	return fmt.Sprintf("%supdate-%s", stng.config.WorkflowTemplatePrefix, stng.suffix)
}

func (stng SuffixedTemplateNameGenerator) DeleteTemplate() string {
	return fmt.Sprintf("%sdelete", stng.config.WorkflowTemplatePrefix)
}

func createProviderServiceUrl(svc corev1.Service, port int) string {
	return fmt.Sprintf("%s.%s:%d", svc.Name, svc.Namespace, port)
}

type ResourceWorkflowFactory[R pipelineshub.Resource, ResourceDefinition any] struct {
	Config                config.KfpControllerConfigSpec
	TemplateNameGenerator TemplateNameGenerator
	DefinitionCreator     func(R) (ResourceDefinition, error)
	WorkflowParamsCreator func(R) ([]argo.Parameter, error)
}

func WorkflowParamsCreatorNoop[R any](_ R) ([]argo.Parameter, error) {
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

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) resourceDefinitionJson(resource R) (string, error) {
	resourceDefinition, err := workflows.DefinitionCreator(resource)
	if err != nil {
		return "", err
	}

	marshalled, err := json.Marshal(&resourceDefinition)
	if err != nil {
		return "", err
	}

	return string(marshalled), nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructCreationWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.resourceDefinitionJson(resource)
	if err != nil {
		return nil, err
	}

	providerConf, err := json.Marshal(provider.Spec)
	if err != nil {
		return nil, err
	}

	namespacedProvider, err := provider.GetCommonNamespacedName().String()
	if err != nil {
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
			Name:  workflowconstants.ProviderConfigParameterName,
			Value: argo.AnyStringPtr(string(providerConf)),
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

	additionalParams, err := workflows.WorkflowParamsCreator(resource)
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
				Name: workflows.TemplateNameGenerator.CreateTemplate(),
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructUpdateWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.resourceDefinitionJson(resource)
	if err != nil {
		return nil, err
	}

	providerConf, err := json.Marshal(provider.Spec)
	if err != nil {
		return nil, err
	}

	namespacedProvider, err := provider.GetCommonNamespacedName().String()
	if err != nil {
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
			Name:  workflowconstants.ProviderConfigParameterName,
			Value: argo.AnyStringPtr(string(providerConf)),
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

	additionalParams, err := workflows.WorkflowParamsCreator(resource)
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
				Name: workflows.TemplateNameGenerator.UpdateTemplate(),
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructDeletionWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	providerConf, err := json.Marshal(provider.Spec)
	if err != nil {
		return nil, err
	}

	namespacedProvider, err := provider.GetCommonNamespacedName().String()
	if err != nil {
		fmt.Println("ResourceWorkflowFactory: err: ", err)
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
						Name:  workflowconstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(string(providerConf)),
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
				Name: workflows.TemplateNameGenerator.DeleteTemplate(),
			},
		},
	}, nil
}
