package pipelines

import (
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha1"
)

var RunConfigurationWorkflowConstants = struct {
	RunConfigurationIdParameterName string
	CreationStepName                string
	DeletionStepName                string
}{
	RunConfigurationIdParameterName: "runconfiguration-id",
	CreationStepName:                "create",
	DeletionStepName:                "delete",
}

type RunConfigurationWorkflowFactory struct {
	WorkflowFactory
}

func (workflows RunConfigurationWorkflowFactory) ConstructCreationWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	creationScriptTemplate, err := workflows.creator(runConfiguration)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: workflows.Config.Argo.ServiceAccount,
			Entrypoint:         WorkflowConstants.EntryPointName,
			Templates: []argo.Template{
				{
					Name: WorkflowConstants.EntryPointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     RunConfigurationWorkflowConstants.CreationStepName,
									Template: RunConfigurationWorkflowConstants.CreationStepName,
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: fmt.Sprintf("{{steps.%s.outputs.result}}",
										RunConfigurationWorkflowConstants.CreationStepName),
								},
							},
						},
					},
				},
				creationScriptTemplate,
			},
		},
	}, nil
}

func (workflows *RunConfigurationWorkflowFactory) ConstructDeletionWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	deletionScriptTemplate, err := workflows.deleter(runConfiguration)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: workflows.Config.Argo.ServiceAccount,
			Entrypoint:         WorkflowConstants.EntryPointName,
			Templates: []argo.Template{
				{
					Name: WorkflowConstants.EntryPointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     RunConfigurationWorkflowConstants.DeletionStepName,
									Template: RunConfigurationWorkflowConstants.DeletionStepName,
								},
							},
						},
					},
				},
				deletionScriptTemplate,
			},
		},
	}, nil
}

func (workflows *RunConfigurationWorkflowFactory) ConstructUpdateWorkflow(runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	deletionScriptTemplate, err := workflows.deleter(runConfiguration)
	if err != nil {
		return nil, err
	}

	creationScriptTemplate, err := workflows.creator(runConfiguration)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(runConfiguration, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: workflows.Config.Argo.ServiceAccount,
			Entrypoint:         WorkflowConstants.EntryPointName,
			Templates: []argo.Template{
				{
					Name: WorkflowConstants.EntryPointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     RunConfigurationWorkflowConstants.DeletionStepName,
									Template: RunConfigurationWorkflowConstants.DeletionStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     RunConfigurationWorkflowConstants.CreationStepName,
									Template: RunConfigurationWorkflowConstants.CreationStepName,
									ContinueOn: &argo.ContinueOn{
										Failed: true,
									},
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: fmt.Sprintf("{{steps.%s.outputs.result}}",
										RunConfigurationWorkflowConstants.CreationStepName),
								},
							},
						},
					},
				},
				deletionScriptTemplate,
				creationScriptTemplate,
			},
		},
	}, nil
}

func (workflows *RunConfigurationWorkflowFactory) creator(runConfiguration *pipelinesv1.RunConfiguration) (argo.Template, error) {
	var experimentName string

	if runConfiguration.Spec.ExperimentName == "" {
		experimentName = workflows.Config.DefaultExperiment
	} else {
		experimentName = runConfiguration.Spec.ExperimentName
	}

	pipelineName, pipelineVersion := runConfiguration.ExtractPipelineNameVersion()

	kfpScript, err := workflows.KfpExt("job submit").
		Param("--experiment-name", experimentName).
		Param("--job-name", runConfiguration.Name).
		Param("--pipeline-name", pipelineName).
		OptParam(("--version-name"), pipelineVersion).
		Param("--cron-expression", runConfiguration.Spec.Schedule).
		Build()

	if err != nil {
		return argo.Template{}, err
	}

	return argo.Template{
		Name:     RunConfigurationWorkflowConstants.CreationStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		Script:   workflows.ScriptTemplate(fmt.Sprintf(`%s | jq -r '."Job Details"."ID"'`, kfpScript)),
	}, nil
}

func (workflows *RunConfigurationWorkflowFactory) deleter(runConfiguration *pipelinesv1.RunConfiguration) (argo.Template, error) {
	kfpScript, err := workflows.KfpExt("job delete").Arg(runConfiguration.Status.KfpId).Build()

	if err != nil {
		return argo.Template{}, err
	}

	succeedOnNotFound := fmt.Sprintf(`KFP_RESULT=$(%s 2>&1) || echo $KFP_RESULT | grep -o 'HTTP response body: {.*}' | cut -d ':' -f 2- | jq -e 'select(.code==5)'`, kfpScript)

	return argo.Template{
		Name:     RunConfigurationWorkflowConstants.DeletionStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		Script:   workflows.ScriptTemplate(succeedOnNotFound),
	}, nil
}
