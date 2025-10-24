package workflowfactory

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowconstants"
	"github.com/sky-uk/kfp-operator/pkg/common"
	providers "github.com/sky-uk/kfp-operator/pkg/providers/base"
)

type PipelineParamsCreator struct{}

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

func (ppc PipelineParamsCreator) pipelineDefinition(
	provider pipelineshub.Provider, pipeline *pipelineshub.Pipeline,
) ([]pipelineshub.Patch, providers.PipelineDefinition, error) {
	framework, found := findFramework(provider, pipeline)

	if !found {
		return nil, providers.PipelineDefinition{}, &workflowconstants.WorkflowParameterError{SubError: fmt.Sprintf("[%s] framework not support by provider", pipeline.Spec.Framework.Name)}
	}

	return framework.Patches, providers.PipelineDefinition{
		Name: common.NamespacedName{
			Namespace: pipeline.ObjectMeta.Namespace,
			Name:      pipeline.ObjectMeta.Name,
		},
		Version:   pipeline.ComputeVersion(),
		Image:     pipeline.Spec.Image,
		Framework: pipeline.Spec.Framework,
		Env:       pipeline.Spec.Env,
	}, nil
}

func (ppc PipelineParamsCreator) additionalParams(provider pipelineshub.Provider, pipeline *pipelineshub.Pipeline) ([]argo.Parameter, error) {
	framework, found := findFramework(provider, pipeline)

	if !found {
		return nil, &workflowconstants.WorkflowParameterError{SubError: fmt.Sprintf("[%s] framework not support by provider", pipeline.Spec.Framework.Name)}
	}

	return []argo.Parameter{
		{
			Name:  workflowconstants.PipelineFrameworkImageParameterName,
			Value: argo.AnyStringPtr(framework.Image),
		},
	}, nil
}

func PipelineWorkflowFactory(
	config config.KfpControllerConfigSpec,
) *ResourceWorkflowFactory[*pipelineshub.Pipeline, providers.PipelineDefinition] {
	creator := PipelineParamsCreator{}
	return NewResourceWorkflowFactory(
		config,
		CompiledSuffix,
		creator.pipelineDefinition,
		creator.additionalParams,
	)
}
