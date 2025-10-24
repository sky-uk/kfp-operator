package workflowfactory

import (
	"encoding/json"
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/jsonutil"
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

type ResourceWorkflowFactory[R pipelineshub.Resource, ResourceDefinition any] struct {
	Config                config.KfpControllerConfigSpec
	TemplateNameGenerator TemplateNameGenerator
	DefinitionCreator     func(pipelineshub.Provider, R) ([]pipelineshub.Patch, ResourceDefinition, error)
	WorkflowParamsCreator func(pipelineshub.Provider, R) ([]argo.Parameter, error)
	workflowBuilder       *BaseWorkflowBuilder
}

func WorkflowParamsCreatorNoop[R any](provider pipelineshub.Provider, _ R) ([]argo.Parameter, error) {
	return []argo.Parameter{}, nil
}

func NewResourceWorkflowFactory[R pipelineshub.Resource, ResourceDefinition any](
	config config.KfpControllerConfigSpec,
	templateNameGenerator TemplateNameGenerator,
	definitionCreator func(pipelineshub.Provider, R) ([]pipelineshub.Patch, ResourceDefinition, error),
	workflowParamsCreator func(pipelineshub.Provider, R) ([]argo.Parameter, error),
) *ResourceWorkflowFactory[R, ResourceDefinition] {
	return &ResourceWorkflowFactory[R, ResourceDefinition]{
		Config:                config,
		TemplateNameGenerator: templateNameGenerator,
		DefinitionCreator:     definitionCreator,
		WorkflowParamsCreator: workflowParamsCreator,
		workflowBuilder:       NewBaseWorkflowBuilder(config),
	}
}

// CommonWorkflowMeta is deprecated, use workflowBuilder.commonWorkflowMeta instead
func (workflows ResourceWorkflowFactory[R, ResourceDefinition]) CommonWorkflowMeta(
	owner pipelineshub.Resource,
) *metav1.ObjectMeta {
	return workflows.workflowBuilder.commonWorkflowMeta(owner)
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

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructCreationWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	resourceDefinition, err := workflows.resourceDefinitionJson(provider, resource)
	if err != nil {
		return nil, err
	}

	baseParams := workflows.workflowBuilder.BuildCreationParams(resourceDefinition)

	additionalParams, err := workflows.WorkflowParamsCreator(provider, resource)
	if err != nil {
		return nil, err
	}

	return workflows.workflowBuilder.BuildWorkflow(
		resource,
		provider,
		providerSvc,
		workflows.TemplateNameGenerator.CreateTemplate(),
		baseParams,
		additionalParams,
	)
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

	baseParams := workflows.workflowBuilder.BuildUpdateParams(
		resourceDefinition,
		resource.GetStatus().Provider.Id,
	)

	additionalParams, err := workflows.WorkflowParamsCreator(provider, resource)
	if err != nil {
		return nil, err
	}

	return workflows.workflowBuilder.BuildWorkflow(
		resource,
		provider,
		providerSvc,
		workflows.TemplateNameGenerator.UpdateTemplate(),
		baseParams,
		additionalParams,
	)
}

func (workflows *ResourceWorkflowFactory[R, ResourceDefinition]) ConstructDeletionWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	baseParams := workflows.workflowBuilder.BuildDeletionParams(resource.GetStatus().Provider.Id)

	return workflows.workflowBuilder.BuildWorkflow(
		resource,
		provider,
		providerSvc,
		workflows.TemplateNameGenerator.DeleteTemplate(),
		baseParams,
		[]argo.Parameter{},
	)
}
