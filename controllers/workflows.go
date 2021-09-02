package controllers

import (
	"gopkg.in/yaml.v2"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	OperationLabelKey = "pipelines.kubeflow.org/operation"
	Create            = "create"
	Update            = "update"
	Delete            = "delete"
)

var pipelineConfigAsYaml = func(pipeline *pipelinesv1.Pipeline) (*argo.AnyString, error) {
	specAsYaml, err := yaml.Marshal(&pipeline.Spec)

	if err != nil {
		return nil, err
	}

	argoYaml := argo.AnyString(specAsYaml)

	return &argoYaml, nil
}

var constructCreationWorkflow = func(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	specAsYaml, err := pipelineConfigAsYaml(pipeline)

	if err != nil {
		return nil, err
	}

	workflow := &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "create-pipeline-" + pipeline.ObjectMeta.Name,
			Namespace: "default",
			Labels: map[string]string{
				OperationLabelKey: Create,
			},
		},
		Spec: argo.WorkflowSpec{
			Entrypoint: "create-pipeline",
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  "config",
						Value: specAsYaml,
					},
				},
			},
			Templates: []argo.Template{
				{
					Name: "create-pipeline",
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: "id",
								ValueFrom: &argo.ValueFrom{
									Parameter: "0",
								},
							},
						},
					},
				},
			},
		},
	}

	return workflow, nil
}

var constructUpdateWorkflow = func(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	specAsYaml, err := pipelineConfigAsYaml(pipeline)

	if err != nil {
		return nil, err
	}

	id := argo.AnyString(pipeline.Status.Id)

	workflow := &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "update-pipeline-" + pipeline.ObjectMeta.Name,
			Namespace: "default",
			Labels: map[string]string{
				OperationLabelKey: Update,
			},
		},
		Spec: argo.WorkflowSpec{
			Entrypoint: "update-pipeline",
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  "pipeline-id",
						Value: &id,
					},
					{
						Name:  "config",
						Value: specAsYaml,
					},
				},
			},
			Templates: []argo.Template{
				{
					Name: "update-pipeline",
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{},
						},
					},
				},
			},
		},
	}

	return workflow, nil
}

var constructDeletionWorkflow = func(pipeline *pipelinesv1.Pipeline) *argo.Workflow {

	id := argo.AnyString(pipeline.Status.Id)

	workflow := &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "delete-pipeline-" + pipeline.ObjectMeta.Name,
			Namespace: "default",
			Labels: map[string]string{
				OperationLabelKey: Delete,
			},
		},
		Spec: argo.WorkflowSpec{
			Entrypoint: "delete-pipeline",
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  "pipeline-id",
						Value: &id,
					},
				},
			},
			Templates: []argo.Template{
				{
					Name: "delete-pipeline",
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{},
						},
					},
				},
			},
		},
	}

	return workflow
}
