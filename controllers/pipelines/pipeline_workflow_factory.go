package pipelines

import (
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
)

type PipelineDefinitionCreator struct {
	Config config.Configuration
}

func (pdc PipelineDefinitionCreator) pipelineDefinition(pipeline *pipelinesv1.Pipeline) (providers.PipelineDefinition, error) {
	return providers.PipelineDefinition{
		Name:          pipeline.ObjectMeta.Name,
		Version:       pipeline.ComputeVersion(),
		Image:         pipeline.Spec.Image,
		TfxComponents: pipeline.Spec.TfxComponents,
		Env:           NamedValuesToMap(pipeline.Spec.Env),
		BeamArgs:      NamedValuesToMultiMap(pipeline.Spec.BeamArgs),
	}, nil
}

func PipelineWorkflowFactory(config config.Configuration) *ResourceWorkflowFactory[*pipelinesv1.Pipeline, providers.PipelineDefinition] {
	return &ResourceWorkflowFactory[*pipelinesv1.Pipeline, providers.PipelineDefinition]{
		DefinitionCreator: PipelineDefinitionCreator{
			Config: config,
		}.pipelineDefinition,
		Config:                config,
		TemplateNameGenerator: CompiledTemplateNameGenerator(config),
	}
}
