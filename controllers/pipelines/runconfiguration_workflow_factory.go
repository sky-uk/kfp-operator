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

func (workflows RunConfigurationWorkflowFactory) providerConfigParameter() argo.Parameter {
	return argo.Parameter{
		Name:  WorkflowConstants.ProviderConfigParameterName,
		Value: argo.AnyStringPtr(workflows.ProviderConfig),
	}
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

func (workflows RunConfigurationWorkflowFactory) ConstructCreationWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	runConfigurationDefinition, err := workflows.runConfigurationDefinitionYaml(runConfiguration)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  RunConfigurationWorkflowConstants.RunConfigurationDefinitionParameterName,
						Value: argo.AnyStringPtr(runConfigurationDefinition),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "create-runconfiguration",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows RunConfigurationWorkflowFactory) ConstructUpdateWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	runConfigurationDefinition, err := workflows.runConfigurationDefinitionYaml(runConfiguration)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  RunConfigurationWorkflowConstants.RunConfigurationDefinitionParameterName,
						Value: argo.AnyStringPtr(runConfigurationDefinition),
					},
					{
						Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
						Value: argo.AnyStringPtr(runConfiguration.Status.ProviderId),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "update-runconfiguration",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows RunConfigurationWorkflowFactory) ConstructDeletionWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
						Value: argo.AnyStringPtr(runConfiguration.Status.ProviderId),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "delete-runconfiguration",
				ClusterScope: true,
			},
		},
	}, nil
}
