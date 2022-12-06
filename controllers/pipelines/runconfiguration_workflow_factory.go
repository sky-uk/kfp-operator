package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/providers/base"
	"gopkg.in/yaml.v2"
)

var RunConfigurationWorkflowConstants = struct {
	RunConfigurationDefinitionParameterName string
	RunConfigurationIdParameterName         string
}{
	RunConfigurationDefinitionParameterName: "runconfiguration-definition",
	RunConfigurationIdParameterName:         "runconfiguration-id",
}

type RunConfigurationWorkflowFactory struct {
	WorkflowFactoryBase
}

func (wf *RunConfigurationWorkflowFactory) runConfigurationDefinitionYaml(runConfiguration *pipelinesv1.RunConfiguration) (string, error) {
	var experimentName string

	if runConfiguration.Spec.ExperimentName == "" {
		experimentName = wf.Config.DefaultExperiment
	} else {
		experimentName = runConfiguration.Spec.ExperimentName
	}

	if runConfiguration.Status.ObservedPipelineVersion == "" {
		return "", fmt.Errorf("unknown pipeline version")
	}

	runConfigurationDefinition := providers.RunConfigurationDefinition{
		Name:            runConfiguration.ObjectMeta.Name,
		Version:         runConfiguration.ComputeVersion(),
		PipelineName:    runConfiguration.Spec.Pipeline.Name,
		PipelineVersion: runConfiguration.Status.ObservedPipelineVersion,
		ExperimentName:  experimentName,
		Schedule:        runConfiguration.Spec.Schedule,
	}

	marshalled, err := yaml.Marshal(&runConfigurationDefinition)

	if err != nil {
		return "", err
	}

	return string(marshalled), nil
}

func (workflows RunConfigurationWorkflowFactory) ConstructCreationWorkflow(provider string, runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	runConfigurationDefinition, err := workflows.runConfigurationDefinitionYaml(runConfiguration)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(runConfiguration, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  RunConfigurationWorkflowConstants.RunConfigurationDefinitionParameterName,
						Value: argo.AnyStringPtr(runConfigurationDefinition),
					},
					{
						Name:  WorkflowConstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.Config.WorkflowTemplatePrefix + "create-runconfiguration",
			},
		},
	}, nil
}

func (workflows RunConfigurationWorkflowFactory) ConstructUpdateWorkflow(provider string, runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	runConfigurationDefinition, err := workflows.runConfigurationDefinitionYaml(runConfiguration)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(runConfiguration, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  RunConfigurationWorkflowConstants.RunConfigurationDefinitionParameterName,
						Value: argo.AnyStringPtr(runConfigurationDefinition),
					},
					{
						Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
						Value: argo.AnyStringPtr(runConfiguration.Status.ProviderId.Id),
					},
					{
						Name:  WorkflowConstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.Config.WorkflowTemplatePrefix + "update-runconfiguration",
			},
		},
	}, nil
}

func (workflows RunConfigurationWorkflowFactory) ConstructDeletionWorkflow(provider string, runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *workflows.CommonWorkflowMeta(runConfiguration, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
						Value: argo.AnyStringPtr(runConfiguration.Status.ProviderId.Id),
					},
					{
						Name:  WorkflowConstants.ProviderNameParameterName,
						Value: argo.AnyStringPtr(provider),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name: workflows.Config.WorkflowTemplatePrefix + "delete-runconfiguration",
			},
		},
	}, nil
}
