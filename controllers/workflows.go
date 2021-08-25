package controllers

import (
	"gopkg.in/yaml.v2"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
)

var constructCreationWorkflow = func(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	specAsYaml, err := yaml.Marshal(&pipeline.Spec)

	if err != nil {
		return nil, err
	}

	configuration := argo.AnyString(specAsYaml)

	workflow := &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "create-pipeline-" + pipeline.ObjectMeta.Name,
			Namespace: "default",
			Labels: map[string]string{
				"operation": "create-pipeline",
			},
		},
		Spec: argo.WorkflowSpec{
			Entrypoint: "create-pipeline",
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  "config",
						Value: &configuration,
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
	specAsYaml, err := yaml.Marshal(&pipeline.Spec)

	if err != nil {
		return nil, err
	}

	configuration := argo.AnyString(specAsYaml)

	workflow := &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "update-pipeline-" + pipeline.ObjectMeta.Name,
			Namespace: "default",
			Labels: map[string]string{
				"operation": "update-pipeline",
			},
		},
		Spec: argo.WorkflowSpec{
			Entrypoint: "update-pipeline",
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  "config",
						Value: &configuration,
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
