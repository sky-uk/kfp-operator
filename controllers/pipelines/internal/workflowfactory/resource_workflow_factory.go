package workflowfactory

import (
	"encoding/json"
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha6"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"gopkg.in/yaml.v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkflowFactory[R pipelinesv1.Resource] interface {
	ConstructCreationWorkflow(provider pipelinesv1.Provider, resource R) (*argo.Workflow, error)
	ConstructUpdateWorkflow(provider pipelinesv1.Provider, resource R) (*argo.Workflow, error)
	ConstructDeletionWorkflow(provider pipelinesv1.Provider, resource R) (*argo.Workflow, error)
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

type ResourceWorkflowFactory[R pipelinesv1.Resource, ResourceDefinition any] struct {
	Config                config.KfpControllerConfigSpec
	TemplateNameGenerator TemplateNameGenerator
	DefinitionCreator     func(R) (ResourceDefinition, error)
}

func (workflows ResourceWorkflowFactory[R, ResourceDefinition]) CommonWorkflowMeta(
	owner pipelinesv1.Resource,
) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: fmt.Sprintf("%s-%s-", owner.GetKind(), owner.GetName()),
		Namespace:    workflows.Config.WorkflowNamespace,
		Labels:       workflowconstants.CommonWorkflowLabels(owner),
	}
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) resourceDefinitionYaml(resource R) (string, error) {
	resourceDefinition, err := workflows.DefinitionCreator(resource)
	if err != nil {
		return "", err
	}

	marshalled, err := yaml.Marshal(&resourceDefinition)
	if err != nil {
		return "", err
	}

	return string(marshalled), nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructCreationWorkflow(
	provider pipelinesv1.Provider,
	resource R,
) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.resourceDefinitionYaml(resource)
	if err != nil {
		return nil, err
	}

	providerConf, err := json.Marshal(provider.Spec)
	if err != nil {
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
						Name:  workflowconstants.ResourceDefinitionParameterName,
						Value: argo.AnyStringPtr(resourceDefinition),
					},
					{
						Name:  workflowconstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider.Name),
					},
					{
						Name:  workflowconstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(string(providerConf)),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.TemplateNameGenerator.CreateTemplate(),
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructUpdateWorkflow(
	provider pipelinesv1.Provider,
	resource R,
) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.resourceDefinitionYaml(resource)
	if err != nil {
		return nil, err
	}

	providerConf, err := json.Marshal(provider.Spec)
	if err != nil {
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
						Name:  workflowconstants.ResourceDefinitionParameterName,
						Value: argo.AnyStringPtr(resourceDefinition),
					},
					{
						Name:  workflowconstants.ResourceIdParameterName,
						Value: argo.AnyStringPtr(resource.GetStatus().Provider.Id),
					},
					{
						Name:  workflowconstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider.Name),
					},
					{
						Name:  workflowconstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(string(providerConf)),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.TemplateNameGenerator.UpdateTemplate(),
			},
		},
	}, nil
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructDeletionWorkflow(
	provider pipelinesv1.Provider,
	resource R,
) (*argo.Workflow, error) {
	providerConf, err := json.Marshal(provider.Spec)
	if err != nil {
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
						Value: argo.AnyStringPtr(provider.Name),
					},
					{
						Name:  workflowconstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(string(providerConf)),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.TemplateNameGenerator.DeleteTemplate(),
			},
		},
	}, nil
}