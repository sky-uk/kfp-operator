package pipelines

import (
	"fmt"
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

var ExperimentWorkflowConstants = struct {
	ExperimentIdParameterName string
	CreationStepName          string
	DeletionStepName          string
	ExperimentKind            string
}{
	ExperimentIdParameterName: "experiment-id",
	CreationStepName:          "create",
	DeletionStepName:          "delete",
	ExperimentKind:            "experiment",
}

type ExperimentWorkflowFactory struct {
	WorkflowFactory
}

func (workflows ExperimentWorkflowFactory) ConstructCreationWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	entrypointName := fmt.Sprintf("%s-%s", WorkflowConstants.CreateOperationLabel, ExperimentWorkflowConstants.ExperimentKind)

	creationScriptTemplate, err := workflows.creator(experiment)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(experiment.NamespacedName(), WorkflowConstants.CreateOperationLabel, ExperimentWorkflowConstants.ExperimentKind),
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
									Name:     ExperimentWorkflowConstants.CreationStepName,
									Template: ExperimentWorkflowConstants.CreationStepName,
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: ExperimentWorkflowConstants.ExperimentIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: fmt.Sprintf("{{steps.%s.outputs.result}}",
										ExperimentWorkflowConstants.CreationStepName),
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

func (workflows *ExperimentWorkflowFactory) ConstructDeletionWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	entrypointName := fmt.Sprintf("%s-%s", WorkflowConstants.DeleteOperationLabel, ExperimentWorkflowConstants.ExperimentKind)

	deletionScriptTemplate, err := workflows.deleter(experiment)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(experiment.NamespacedName(), WorkflowConstants.DeleteOperationLabel, ExperimentWorkflowConstants.ExperimentKind),
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
									Name:     ExperimentWorkflowConstants.DeletionStepName,
									Template: ExperimentWorkflowConstants.DeletionStepName,
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

func (workflows *ExperimentWorkflowFactory) ConstructUpdateWorkflow(experiment *pipelinesv1.Experiment) (*argo.Workflow, error) {
	entrypointName := fmt.Sprintf("%s-%s", WorkflowConstants.UpdateOperationLabel, ExperimentWorkflowConstants.ExperimentKind)

	deletionScriptTemplate, err := workflows.deleter(experiment)
	if err != nil {
		return nil, err
	}

	creationScriptTemplate, err := workflows.creator(experiment)
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(experiment.NamespacedName(), WorkflowConstants.UpdateOperationLabel, ExperimentWorkflowConstants.ExperimentKind),
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
									Name:     ExperimentWorkflowConstants.DeletionStepName,
									Template: ExperimentWorkflowConstants.DeletionStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     ExperimentWorkflowConstants.CreationStepName,
									Template: ExperimentWorkflowConstants.CreationStepName,
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
								Name: ExperimentWorkflowConstants.ExperimentIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: fmt.Sprintf("{{steps.%s.outputs.result}}",
										ExperimentWorkflowConstants.CreationStepName),
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

func (workflows *ExperimentWorkflowFactory) creator(experiment *pipelinesv1.Experiment) (argo.Template, error) {
	kfpScript, err := workflows.KfpExt("experiment create").
		OptParam("--description", experiment.Spec.Description).
		Arg(experiment.Name).
		Build()

	if err != nil {
		return argo.Template{}, err
	}

	return argo.Template{
		Name:     ExperimentWorkflowConstants.CreationStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		Script:   workflows.ScriptTemplate(fmt.Sprintf(`%s | jq -r '."ID"'`, kfpScript)),
	}, nil
}

func (workflows *ExperimentWorkflowFactory) deleter(experiment *pipelinesv1.Experiment) (argo.Template, error) {
	kfpScript, err := workflows.KfpExt("experiment delete").Arg(experiment.Status.KfpId).Build()

	if err != nil {
		return argo.Template{}, err
	}

	return argo.Template{
		Name:     ExperimentWorkflowConstants.DeletionStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		// The KFP SDK requires confirmation of the deletion and does not provide a flag to circumnavigate this
		Script: workflows.ScriptTemplate(fmt.Sprintf("echo y | %s", kfpScript)),
	}, nil
}
