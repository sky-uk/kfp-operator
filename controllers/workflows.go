package controllers

import (
	"fmt"

	"gopkg.in/yaml.v2"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	OperationLabelKey         = "pipelines.kubeflow.org/operation"
	PipelineLabelKey          = "pipelines.kubeflow.org/pipeline"
	PipelineConfigKey         = "pipeline-config"
	PipelineIdParameterName   = "pipeline-id"
	PipelineYamlParameterName = "pipeline"
	Create                    = "create-pipeline"
	Update                    = "update-pipeline"
	Delete                    = "delete-pipeline"
	CompileStepName           = "compile"
	UploadStepName            = "upload"
	Namespace                 = "default"
	PipelineYamlFilePath      = "/tmp/pipeline.yaml"
	PipelineIdFilePath        = "/tmp/pipeline.txt"
)

var (
	trueValue = true
)

type WorkflowConfiguration struct {
	CompilerImage string
	UploaderImage string
}

type Workflows struct {
	Config WorkflowConfiguration
}

var pipelineConfigAsYaml = func(pipelineSpec *pipelinesv1.PipelineSpec) (string, error) {
	specAsYaml, err := yaml.Marshal(&pipelineSpec)

	if err != nil {
		return "", err
	}

	return string(specAsYaml), nil
}

func commonMeta(pipeline *pipelinesv1.Pipeline, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-",
		Namespace:    Namespace,
		Labels: map[string]string{
			OperationLabelKey: operation,
			PipelineLabelKey:  pipeline.ObjectMeta.Name,
		},
	}
}

func parameter(key string, value string) *argo.Parameter {
	return &argo.Parameter{
		Name:  key,
		Value: argo.AnyStringPtr(value),
	}
}

func pipelineIdParameter(id string) *argo.Parameter {
	return parameter(PipelineIdParameterName, id)
}

func pipelineConfigParameter(pipelineSpec *pipelinesv1.PipelineSpec) (*argo.Parameter, error) {
	specAsYaml, err := pipelineConfigAsYaml(pipelineSpec)

	if err != nil {
		return nil, err
	}

	return parameter(PipelineConfigKey, specAsYaml), nil
}

func (workflows Workflows) ConstructCreationWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	yamlConfig, error := pipelineConfigAsYaml(&pipeline.Spec)

	if error != nil {
		return nil, error
	}

	entrypointName := Create

	workflow := &argo.Workflow{
		ObjectMeta: *commonMeta(pipeline, Create),
		Spec: argo.WorkflowSpec{
			Entrypoint: entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     CompileStepName,
									Template: CompileStepName,
								},
								{
									Name:     UploadStepName,
									Template: UploadStepName,
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: PipelineIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: "steps.upload.outputs.result",
								},
							},
						},
					},
				},
				workflows.compiler(yamlConfig),
				workflows.uploader(&pipeline.Spec, pipeline.Status.Version),
			},
		},
	}

	return workflow, nil
}

func (workflows *Workflows) compiler(pipelineSpec string) argo.Template {
	compilerVolumeName := "compiler"
	compilerVolumePath := "/compiler"

	args := []string{
		"compile",
		fmt.Sprintf("--pipeline_config=%s", pipelineSpec),
	}

	return argo.Template{
		Name: CompileStepName,
		Outputs: argo.Outputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineYamlParameterName,
					Path: PipelineYamlFilePath,
				},
			},
		},
		Container: &apiv1.Container{
			Name:  CompileStepName,
			Image: workflows.Config.CompilerImage,
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      compilerVolumeName,
					MountPath: compilerVolumePath,
				},
			},
			Args: args,
		},
		InitContainers: []argo.UserContainer{
			{
				Container:          apiv1.Container{},
				MirrorVolumeMounts: &trueValue,
			},
		},
		Volumes: []apiv1.Volume{
			{
				Name: compilerVolumeName,
			},
		},
	}
}

func (workflows *Workflows) uploader(pipelineSpec *pipelinesv1.PipelineSpec, version string) argo.Template {

	args := []string{
		"create",
		fmt.Sprintf("--input_file=%s", PipelineYamlFilePath),
		fmt.Sprintf("--version=%s", version),
		fmt.Sprintf("--output_file=%s", PipelineIdFilePath),
	}

	return argo.Template{
		Name: UploadStepName,
		Outputs: argo.Outputs{
			Parameters: []argo.Parameter{
				{
					Name: PipelineIdParameterName,
					ValueFrom: &argo.ValueFrom{
						Path: PipelineIdFilePath,
					},
				},
			},
		},
		Container: &apiv1.Container{
			Name:  UploadStepName,
			Image: workflows.Config.UploaderImage,
			Args:  args,
		},
	}
}

func (workflows Workflows) ConstructUpdateWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	configParameter, error := pipelineConfigParameter(&pipeline.Spec)

	if error != nil {
		return nil, error
	}

	workflow := &argo.Workflow{
		ObjectMeta: *commonMeta(pipeline, Update),
		Spec: argo.WorkflowSpec{
			Entrypoint: Update,
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					*pipelineIdParameter(pipeline.Status.Id),
					*configParameter,
				},
			},
			Templates: []argo.Template{
				{
					Name: Update,
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

func (workflows Workflows) ConstructDeletionWorkflow(pipeline *pipelinesv1.Pipeline) *argo.Workflow {

	workflow := &argo.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: Delete + "-",
			Namespace:    Namespace,
			Labels: map[string]string{
				OperationLabelKey: Delete,
				PipelineLabelKey:  pipeline.ObjectMeta.Name,
			},
		},
		Spec: argo.WorkflowSpec{
			Entrypoint: Delete,
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					*pipelineIdParameter(pipeline.Status.Id),
				},
			},
			Templates: []argo.Template{
				{
					Name: Delete,
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
