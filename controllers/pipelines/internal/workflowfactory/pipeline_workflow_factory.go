package workflowfactory

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/samber/lo"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"github.com/sky-uk/kfp-operator/internal/config"
	"github.com/sky-uk/kfp-operator/pkg/common"
	providers "github.com/sky-uk/kfp-operator/pkg/providers/base"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func findFramework(provider pipelineshub.Provider, pipeline *pipelineshub.Pipeline) (*pipelineshub.Framework, bool) {
	requestedFramework := strings.ToLower(pipeline.Spec.Framework.Name)

	framework, ok := lo.Find(provider.Spec.Frameworks, func(framework pipelineshub.Framework) bool {
		return framework.Name == requestedFramework
	})

	if !ok {
		return nil, false
	}

	return &framework, true
}

func pipelineDefinition(pipeline *pipelineshub.Pipeline) providers.PipelineDefinition {
	return providers.PipelineDefinition{
		Name: common.NamespacedName{
			Namespace: pipeline.ObjectMeta.Namespace,
			Name:      pipeline.ObjectMeta.Name,
		},
		Version:   pipeline.ComputeVersion(),
		Image:     pipeline.Spec.Image,
		Framework: pipeline.Spec.Framework,
		Env:       pipeline.Spec.Env,
	}
}

// pipelineWorkflowFactory satisfies WorkflowFactory for pipelines, whose
// provider definition depends on the framework named by the pipeline. The
// provider supplies that framework's definition patches and image parameter.
type pipelineWorkflowFactory struct {
	assembler workflowAssembler
}

// creationParams returns the argo parameters for creating or updating a
// pipeline: its patched definition and the image of the framework it names.
func (f pipelineWorkflowFactory) creationParams(
	provider pipelineshub.Provider,
	pipeline *pipelineshub.Pipeline,
) ([]argo.Parameter, error) {
	framework, found := findFramework(provider, pipeline)
	if !found {
		return nil, &workflowconstants.WorkflowParameterError{
			SubError: fmt.Sprintf("[%s] framework not support by provider", pipeline.Spec.Framework.Name),
		}
	}

	definitionJson, err := marshalDefinition(pipelineDefinition(pipeline), framework.Patches)
	if err != nil {
		return nil, err
	}

	return []argo.Parameter{
		definitionParam(definitionJson),
		{
			Name:  workflowconstants.PipelineFrameworkImageParameterName,
			Value: argo.AnyStringPtr(framework.Image),
		},
	}, nil
}

func (f pipelineWorkflowFactory) ConstructCreationWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	pipeline *pipelineshub.Pipeline,
) (*argo.Workflow, error) {
	params, err := f.creationParams(provider, pipeline)
	if err != nil {
		return nil, err
	}

	return f.assembler.constructWorkflow(
		provider,
		providerSvc,
		pipeline,
		f.assembler.createTemplateName(CompiledSuffix),
		params,
	)
}

func (f pipelineWorkflowFactory) ConstructUpdateWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	pipeline *pipelineshub.Pipeline,
) (*argo.Workflow, error) {
	params, err := f.creationParams(provider, pipeline)
	if err != nil {
		return nil, err
	}

	return f.assembler.constructWorkflow(
		provider,
		providerSvc,
		pipeline,
		f.assembler.updateTemplateName(CompiledSuffix),
		append(params, resourceIdParam(pipeline)),
	)
}

func (f pipelineWorkflowFactory) ConstructDeletionWorkflow(
	provider pipelineshub.Provider,
	providerSvc corev1.Service,
	pipeline *pipelineshub.Pipeline,
) (*argo.Workflow, error) {
	return f.assembler.constructWorkflow(
		provider,
		providerSvc,
		pipeline,
		f.assembler.deleteTemplateName(),
		[]argo.Parameter{resourceIdParam(pipeline)},
	)
}

func PipelineWorkflowFactory(
	config config.ConfigSpec,
) WorkflowFactory[*pipelineshub.Pipeline] {
	return pipelineWorkflowFactory{
		assembler: workflowAssembler{config: config},
	}
}
