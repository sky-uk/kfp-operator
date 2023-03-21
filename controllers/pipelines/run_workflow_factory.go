package pipelines

import (
	"fmt"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha4"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
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

	if run.Status.ObservedPipelineVersion == "" {
		return providers.RunDefinition{}, fmt.Errorf("unknown pipeline version")
	}

	return providers.RunDefinition{
		Name:              common.NamespacedName{Name: run.Name, Namespace: run.Namespace},
		Version:           run.ComputeVersion(),
		PipelineName:      run.Spec.Pipeline.Name,
		PipelineVersion:   run.Status.ObservedPipelineVersion,
		ExperimentName:    experimentName,
		RuntimeParameters: NamedValuesToMap(run.Spec.RuntimeParameters),
	}, nil
}

func RunWorkflowFactory(config config.Configuration) ResourceWorkflowFactory[*pipelinesv1.Run, providers.RunDefinition] {
	return ResourceWorkflowFactory[*pipelinesv1.Run, providers.RunDefinition]{
		DefinitionCreator: RunDefinitionCreator{
			Config: config,
		}.runDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
	}
}
