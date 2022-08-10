package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha2"
)

var RunConfigurationWorkflowConstants = struct {
	JobIdParameterName           string
	JobNameParameterName         string
	PipelineNameParameterName    string
	PipelineVersionParameterName string
	ExperimentNameParameterName  string
	ScheduleParameterName        string
}{
	JobIdParameterName:           "job-id",
	JobNameParameterName:         "job-name",
	PipelineNameParameterName:    "pipeline-name",
	PipelineVersionParameterName: "pipeline-version",
	ExperimentNameParameterName:  "experiment-name",
	ScheduleParameterName:        "schedule",
}

type RunConfigurationWorkflowFactory struct {
	WorkflowFactoryBase
}

func (workflows RunConfigurationWorkflowFactory) ConstructCreationWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	if runConfiguration.Status.ObservedPipelineVersion == "" {
		return nil, fmt.Errorf("unknown pipeline version")
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: workflows.creationParameters(runConfiguration),
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         "create-runconfiguration",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows *RunConfigurationWorkflowFactory) ConstructUpdateWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: append(workflows.creationParameters(runConfiguration), workflows.deletionParameters(runConfiguration)...),
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         "update-runconfiguration",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows *RunConfigurationWorkflowFactory) ConstructDeletionWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: workflows.deletionParameters(runConfiguration),
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         "delete-runconfiguration",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows *RunConfigurationWorkflowFactory) deletionParameters(runConfiguration *pipelinesv1.RunConfiguration) []argo.Parameter {
	return []argo.Parameter{
		{
			Name:  RunConfigurationWorkflowConstants.JobIdParameterName,
			Value: argo.AnyStringPtr(runConfiguration.Status.KfpId),
		},
	}
}

func (workflows *RunConfigurationWorkflowFactory) creationParameters(runConfiguration *pipelinesv1.RunConfiguration) []argo.Parameter {
	var experimentName string
	if runConfiguration.Spec.ExperimentName == "" {
		experimentName = workflows.Config.DefaultExperiment
	} else {
		experimentName = runConfiguration.Spec.ExperimentName
	}

	return []argo.Parameter{
		{
			Name:  RunConfigurationWorkflowConstants.JobIdParameterName,
			Value: argo.AnyStringPtr(runConfiguration.Status.KfpId),
		},
		{
			Name:  RunConfigurationWorkflowConstants.JobNameParameterName,
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
	}
}
