package pipelines

import (
	"fmt"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
)

type RunDefinitionCreator struct {
	Config config.KfpControllerConfigSpec
}

func (rdc RunDefinitionCreator) runDefinition(run *pipelinesv1.Run) (providers.RunDefinition, error) {
	var experimentName common.NamespacedName
	if run.Spec.ExperimentName == "" {
		experimentName = common.NamespacedName{Name: rdc.Config.DefaultExperiment}
	} else {
		experimentName = common.NamespacedName{Name: run.Spec.ExperimentName, Namespace: run.Namespace}
	}

	if run.Status.ObservedPipelineVersion == "" {
		return providers.RunDefinition{}, fmt.Errorf("unknown pipeline version")
	}

	runtimeParameters, err := run.Spec.ResolveRuntimeParameters(run.Status.Dependencies)
	if err != nil {
		return providers.RunDefinition{}, err
	}

	runDefinition := providers.RunDefinition{
		Name:              common.NamespacedName{Namespace: run.Namespace, Name: run.Name},
		Version:           run.ComputeVersion(),
		PipelineName:      common.NamespacedName{Namespace: run.Namespace, Name: run.Spec.Pipeline.Name},
		PipelineVersion:   run.Status.ObservedPipelineVersion,
		ExperimentName:    experimentName,
		RuntimeParameters: NamedValuesToMap(runtimeParameters),
		Artifacts:         run.Spec.Artifacts,
	}

	if runConfigurationName, ok := run.Labels[RunConfigurationConstants.RunConfigurationNameLabelKey]; ok {
		runDefinition.RunConfigurationName = common.NamespacedName{Namespace: run.Namespace, Name: runConfigurationName}
	}

	return runDefinition, nil
}

func RunWorkflowFactory(config config.KfpControllerConfigSpec) *ResourceWorkflowFactory[*pipelinesv1.Run, providers.RunDefinition] {
	return &ResourceWorkflowFactory[*pipelinesv1.Run, providers.RunDefinition]{
		DefinitionCreator: RunDefinitionCreator{
			Config: config,
		}.runDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
	}
}
