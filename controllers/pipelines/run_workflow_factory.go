package pipelines

import (
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/providers/base"
)

type RunDefinitionCreator struct {
	Config config.Configuration
}

func (rdc RunDefinitionCreator) runDefinition(run *pipelinesv1.Run) (providers.RunDefinition, error) {
	var experimentName string

	if run.Spec.ExperimentName == "" {
		experimentName = rdc.Config.DefaultExperiment
	} else {
		experimentName = run.Spec.ExperimentName
	}

	return providers.RunDefinition{
		Name:              run.ObjectMeta.Name,
		Version:           run.ComputeVersion(),
		PipelineName:      run.Spec.Pipeline.Name,
		ExperimentName:    experimentName,
		RuntimeParameters: NamedValuesToMap(run.Spec.RuntimeParameters),
	}, nil
}

func RunWorkflowFactory(config config.Configuration) WorkflowFactory[*pipelinesv1.Run] {
	return &ResourceWorkflowFactory[*pipelinesv1.Run, providers.RunDefinition]{
		DefinitionCreator: RunDefinitionCreator{
			Config: config,
		}.runDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
	}
}
