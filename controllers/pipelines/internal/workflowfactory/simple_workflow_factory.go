package workflowfactory

import (
	"encoding/json"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	corev1 "k8s.io/api/core/v1"
)

// simpleWorkflowFactory satisfies WorkflowFactory for resources whose provider
// definition is derived solely from the resource itself, with no patches or
// framework-specific parameters (Run, RunSchedule, Experiment).
type simpleWorkflowFactory[R pipelineshub.Resource, D any] struct {
	assembler workflowAssembler
	builder   definitionBuilder[R, D]
}

func (f simpleWorkflowFactory[R, D]) definitionParam(resource R) (argo.Parameter, error) {
	definition, err := f.builder.build(resource)
	if err != nil {
		return argo.Parameter{}, err
	}

	definitionJson, err := json.Marshal(definition)
	if err != nil {
		return argo.Parameter{}, err
	}

	return definitionParam(string(definitionJson)), nil
}

func (f simpleWorkflowFactory[R, D]) ConstructCreationWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	definition, err := f.definitionParam(resource)
	if err != nil {
		return nil, err
	}

	return f.assembler.constructWorkflow(
		provider,
		providerSvc,
		resource,
		f.assembler.createTemplateName(SimpleSuffix),
		[]argo.Parameter{definition},
	)
}

func (f simpleWorkflowFactory[R, D]) ConstructUpdateWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	definition, err := f.definitionParam(resource)
	if err != nil {
		return nil, err
	}

	return f.assembler.constructWorkflow(
		provider,
		providerSvc,
		resource,
		f.assembler.updateTemplateName(SimpleSuffix),
		[]argo.Parameter{definition, resourceIdParam(resource)},
	)
}

func (f simpleWorkflowFactory[R, D]) ConstructDeletionWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	resource R,
) (*argo.Workflow, error) {
	return f.assembler.constructWorkflow(
		provider,
		providerSvc,
		resource,
		f.assembler.deleteTemplateName(),
		[]argo.Parameter{resourceIdParam(resource)},
	)
}
