package pipelines

import (
	"fmt"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/providers/base"
)

type RunConfigurationDefinitionCreator struct {
	Config config.Configuration
}

func (rcdc RunConfigurationDefinitionCreator) runConfigurationDefinition(runConfiguration *pipelinesv1.RunConfiguration) (providers.RunConfigurationDefinition, error) {
	var experimentName string

	if runConfiguration.Spec.ExperimentName == "" {
		experimentName = rcdc.Config.DefaultExperiment
	} else {
		experimentName = runConfiguration.Spec.ExperimentName
	}

	if runConfiguration.Status.ObservedPipelineVersion == "" {
		return providers.RunConfigurationDefinition{}, fmt.Errorf("unknown pipeline version")
	}

	return providers.RunConfigurationDefinition{
		Name:              runConfiguration.ObjectMeta.Name,
		Version:           runConfiguration.ComputeVersion(),
		PipelineName:      runConfiguration.Spec.Pipeline.Name,
		PipelineVersion:   runConfiguration.Status.ObservedPipelineVersion,
		ExperimentName:    experimentName,
		Schedule:          runConfiguration.Spec.Schedule,
		RuntimeParameters: NamedValuesToMap(runConfiguration.Spec.RuntimeParameters),
	}, nil
}

func RunConfigurationWorkflowFactory(config config.Configuration) WorkflowFactory[*pipelinesv1.RunConfiguration] {
	return &ResourceWorkflowFactory[*pipelinesv1.RunConfiguration, providers.RunConfigurationDefinition]{
		DefinitionCreator: RunConfigurationDefinitionCreator{
			Config: config,
		}.runConfigurationDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
	}
}
