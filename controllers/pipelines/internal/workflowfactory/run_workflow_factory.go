package workflowfactory

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis"
	config "github.com/sky-uk/kfp-operator/apis/config/hub"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
)

var RunConfigurationConstants = struct {
	RunConfigurationNameLabelKey string
}{
	RunConfigurationNameLabelKey: apis.Group + "/runconfiguration.name",
}

type RunDefinitionCreator struct {
	Config config.KfpControllerConfigSpec
}

func (rdc RunDefinitionCreator) runDefinition(run *pipelineshub.Run) (providers.RunDefinition, error) {
	var experimentName common.NamespacedName
	if run.Spec.ExperimentName == "" {
		experimentName = common.NamespacedName{
			Name: rdc.Config.DefaultExperiment,
		}
	} else {
		experimentName = common.NamespacedName{
			Name:      run.Spec.ExperimentName,
			Namespace: run.Namespace,
		}
	}

	if run.Status.Dependencies.Pipeline.Version == "" {
		return providers.RunDefinition{}, fmt.Errorf("unknown pipeline version")
	}

	parameters, err := run.Spec.ResolveParameters(run.Status.Dependencies)
	if err != nil {
		return providers.RunDefinition{}, err
	}

	runDefinition := providers.RunDefinition{
		Name: common.NamespacedName{
			Namespace: run.Namespace,
			Name:      run.Name,
		},
		Version: run.ComputeVersion(),
		PipelineName: common.NamespacedName{
			Namespace: run.Namespace,
			Name:      run.Spec.Pipeline.Name,
		},
		PipelineVersion: run.Status.Dependencies.Pipeline.Version,
		ExperimentName:  experimentName,
		Parameters:      NamedValuesToMap(parameters),
		Artifacts:       run.Spec.Artifacts,
	}

	if runConfigurationName, ok := run.Labels[RunConfigurationConstants.RunConfigurationNameLabelKey]; ok {
		runDefinition.RunConfigurationName = common.NamespacedName{
			Namespace: run.Namespace,
			Name:      runConfigurationName,
		}
	}

	return runDefinition, nil
}

func RunWorkflowFactory(
	config config.KfpControllerConfigSpec,
) *ResourceWorkflowFactory[*pipelineshub.Run, providers.RunDefinition] {
	return &ResourceWorkflowFactory[*pipelineshub.Run, providers.RunDefinition]{
		DefinitionCreator: RunDefinitionCreator{
			Config: config,
		}.runDefinition,
		Config:                config,
		TemplateNameGenerator: SimpleTemplateNameGenerator(config),
		WorkflowParamsCreator: WorkflowParamsCreatorNoop[*pipelineshub.Run],
	}
}
