package pipelines

import (
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/providers/base"
)

type ExperimentDefinitionCreator struct {
	Config config.Configuration
}

func (edc ExperimentDefinitionCreator) experimentDefinition(experiment *pipelinesv1.Experiment) (providers.ExperimentDefinition, error) {
	return providers.ExperimentDefinition{
		Name:        experiment.ObjectMeta.Name,
		Version:     experiment.ComputeVersion(),
		Description: experiment.Spec.Description,
	}, nil
}

func ExperimentWorkflowFactory(config config.Configuration) WorkflowFactory[*pipelinesv1.Experiment] {
	return &ResourceWorkflowFactory[*pipelinesv1.Experiment, providers.ExperimentDefinition]{
		DefinitionCreator: ExperimentDefinitionCreator{
			Config: config,
		}.experimentDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
	}
}
