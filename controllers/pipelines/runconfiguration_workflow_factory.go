package pipelines

import (
	"context"
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (workflows *RunConfigurationWorkflowFactory) commonMeta(_ context.Context, rc *pipelinesv1.RunConfiguration, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-",
		Namespace:    rc.GetNamespace(),
		Labels:       workflows.Labels(rc, operation),
	}
}

func (workflows *RunConfigurationWorkflowFactory) Labels(resource Resource, operation string) map[string]string {
	return map[string]string{
		WorkflowConstants.OperationLabelKey: operation,
		WorkflowConstants.OwnerKindLabelKey: "runconfiguration",
		WorkflowConstants.OwnerNameLabelKey: resource.GetName(),
	}
}

func (workflows RunConfigurationWorkflowFactory) ConstructCreationWorkflow(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	entrypointName := WorkflowConstants.OperationLabelKey

	creationScriptTemplate, err := workflows.creator(runConfiguration)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.commonMeta(ctx, runConfiguration, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: workflows.Config.Argo.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
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

func (workflows *RunConfigurationWorkflowFactory) ConstructDeletionWorkflow(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	entrypointName := WorkflowConstants.DeleteOperationLabel

	deletionScriptTemplate, err := workflows.deleter(runConfiguration)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.commonMeta(ctx, runConfiguration, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: workflows.Config.Argo.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
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

func (workflows *RunConfigurationWorkflowFactory) ConstructUpdateWorkflow(ctx context.Context, runConfiguration *pipelinesv1.RunConfiguration) (*argo.Workflow, error) {
	entrypointName := WorkflowConstants.UpdateOperationLabel

	deletionScriptTemplate, err := workflows.deleter(runConfiguration)
	if err != nil {
		return nil, err
	}

	creationScriptTemplate, err := workflows.creator(runConfiguration)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *workflows.commonMeta(ctx, runConfiguration, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: workflows.Config.Argo.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
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

	kfpScript, err := workflows.KfpExt("job submit").
		Param("--experiment-name", experimentName).
		Param("--job-name", runConfiguration.Name).
		Param("--pipeline-name", runConfiguration.Spec.PipelineName).
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

	return argo.Template{
		Name:     RunConfigurationWorkflowConstants.DeletionStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		Script:   workflows.ScriptTemplate(kfpScript),
	}, nil
}
