package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
)

var RunConfigurationWorkflowConstants = struct {
	RunConfigurationIdParameterName   string
	RunConfigurationNameParameterName string
	PipelineNameParameterName         string
	PipelineVersionParameterName      string
	ExperimentNameParameterName       string
	ScheduleParameterName             string
}{
	RunConfigurationIdParameterName:   "runconfiguration-id",
	RunConfigurationNameParameterName: "runconfiguration-name",
	PipelineNameParameterName:         "pipeline-name",
	PipelineVersionParameterName:      "pipeline-version",
	ExperimentNameParameterName:       "experiment-name",
	ScheduleParameterName:             "schedule",
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

func (workflows RunConfigurationWorkflowFactory) ConstructCreationWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	creationParameters, err := workflows.creationParameters(runConfiguration)

	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: append(creationParameters, workflows.providerConfigParameter()),
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "create-runconfiguration",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows RunConfigurationWorkflowFactory) ConstructUpdateWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	deletionParameters := workflows.deletionParameters(runConfiguration)
	creationParameters, err := workflows.creationParameters(runConfiguration)

	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: append(append(deletionParameters, workflows.providerConfigParameter()), creationParameters...),
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
				Parameters: append(workflows.deletionParameters(runConfiguration), workflows.providerConfigParameter()),
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "delete-runconfiguration",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows RunConfigurationWorkflowFactory) deletionParameters(runConfiguration *pipelinesv1.RunConfiguration) []argo.Parameter {
	return []argo.Parameter{
		{
			Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
			Value: argo.AnyStringPtr(runConfiguration.Status.KfpId),
		},
	}
}

func (workflows RunConfigurationWorkflowFactory) creationParameters(runConfiguration *pipelinesv1.RunConfiguration) ([]argo.Parameter, error) {
	var experimentName string
	if runConfiguration.Spec.ExperimentName == "" {
		experimentName = workflows.Config.DefaultExperiment
	} else {
		experimentName = runConfiguration.Spec.ExperimentName
	}

	if runConfiguration.Status.ObservedPipelineVersion == "" {
		return nil, fmt.Errorf("unknown pipeline version")
	}

	return []argo.Parameter{
		{
			Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
			Value: argo.AnyStringPtr(runConfiguration.Status.KfpId),
		},
		{
			Name:  RunConfigurationWorkflowConstants.RunConfigurationNameParameterName,
			Value: argo.AnyStringPtr(runConfiguration.Name),
		},
		{
			Name:  RunConfigurationWorkflowConstants.ExperimentNameParameterName,
			Value: argo.AnyStringPtr(experimentName),
		},
		{
			Name:  RunConfigurationWorkflowConstants.PipelineNameParameterName,
			Value: argo.AnyStringPtr(runConfiguration.Spec.Pipeline.Name),
		},
		{
			Name:  RunConfigurationWorkflowConstants.PipelineVersionParameterName,
			Value: argo.AnyStringPtr(runConfiguration.Status.ObservedPipelineVersion),
		},
		{
			Name:  RunConfigurationWorkflowConstants.ScheduleParameterName,
			Value: argo.AnyStringPtr(runConfiguration.Spec.Schedule),
		},
	}, nil
}
