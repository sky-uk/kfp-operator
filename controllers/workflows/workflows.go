package workflows

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
	DeletionStepName          = "delete"
	UpdateStepName            = "update"
	PipelineYamlFilePath      = "/tmp/pipeline.yaml"
	PipelineIdFilePath        = "/tmp/pipeline.txt"
)

var (
	trueValue = true
)

type Configuration struct {
	ImagePullPolicy string
	Namespace       string
	KfpEndpoint     string
	CompilerImage   string
	KfpToolsImage   string
}

type Workflows struct {
	Config Configuration
}

var pipelineConfigAsYaml = func(pipeline *pipelinesv1.Pipeline) (string, error) {
	config := &CompilerConfig{
		Spec:         pipeline.Spec,
		Name:         pipeline.Name,
		PipelineRoot: "root",
		ServingDir:   "serving",
	}

	specAsYaml, err := yaml.Marshal(&config)

	if err != nil {
		return "", err
	}

	return string(specAsYaml), nil
}

func (w *Workflows) commonMeta(pipeline *pipelinesv1.Pipeline, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-",
		Namespace:    w.Config.Namespace,
		Labels: map[string]string{
			OperationLabelKey: operation,
			PipelineLabelKey:  pipeline.ObjectMeta.Name,
		},
	}
}

func (w Workflows) ConstructCreationWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	yamlConfig, error := pipelineConfigAsYaml(pipeline)

	if error != nil {
		return nil, error
	}

	entrypointName := Create

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(pipeline, Create),
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
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     UploadStepName,
									Template: UploadStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: PipelineYamlParameterName,
												From: "{{steps.compile.outputs.artifacts.pipeline}}",
											},
										},
									},
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     UpdateStepName,
									Template: UpdateStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: PipelineYamlParameterName,
												From: "{{steps.compile.outputs.artifacts.pipeline}}",
											},
										},
										Parameters: []argo.Parameter{
											{
												Name:  PipelineIdParameterName,
												Value: argo.AnyStringPtr("{{steps.upload.outputs.result}}"),
											},
										},
									},
								},
							},
						},
					},
					Outputs: argo.Outputs{
						Parameters: []argo.Parameter{
							{
								Name: PipelineIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: "{{steps.upload.outputs.result}}",
								},
							},
						},
					},
				},
				w.compiler(yamlConfig, pipeline.Spec.Image),
				w.uploader(pipeline.Name),
				w.updater(pipeline.Status.Version),
			},
		},
	}

	return workflow, nil
}

func (w Workflows) ConstructUpdateWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	yamlConfig, error := pipelineConfigAsYaml(pipeline)

	if error != nil {
		return nil, error
	}

	entrypointName := Update

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(pipeline, Update),
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
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     UpdateStepName,
									Template: UpdateStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: PipelineYamlParameterName,
												From: "{{steps.compile.outputs.artifacts.pipeline}}",
											},
										},
										Parameters: []argo.Parameter{
											{
												Name:  PipelineIdParameterName,
												Value: argo.AnyStringPtr(pipeline.Status.Id),
											},
										},
									},
								},
							},
						},
					},
				},
				w.compiler(yamlConfig, pipeline.Spec.Image),
				w.updater(pipeline.Status.Version),
			},
		},
	}

	return workflow, nil
}

func (w Workflows) ConstructDeletionWorkflow(pipeline *pipelinesv1.Pipeline) *argo.Workflow {

	entrypointName := Delete

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(pipeline, Delete),
		Spec: argo.WorkflowSpec{
			Entrypoint: entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     DeletionStepName,
									Template: DeletionStepName,
									Arguments: argo.Arguments{
										Parameters: []argo.Parameter{
											{
												Name:  PipelineIdParameterName,
												Value: argo.AnyStringPtr(pipeline.Status.Id),
											},
										},
									},
								},
							},
						},
					},
				},
				w.deleter(),
			},
		},
	}

	return workflow
}

type CompilerConfig struct {
	Spec         pipelinesv1.PipelineSpec
	Name         string
	ServingDir   string            `yaml:"servingDir"`
	PipelineRoot string            `yaml:"pipelineRoot"`
	BeamArgs     map[string]string `yaml:"beamArgs"`
}

func (workflows *Workflows) compiler(pipelineSpec string, pipelineImage string) argo.Template {
	compilerVolumeName := "compiler"
	compilerVolumePath := "/compiler"

	args := []string{
		"/compiler/compiler.py",
		"--output_file",
		PipelineYamlFilePath,
		"--pipeline_config",
		pipelineSpec,
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
			Name:            "pipeline",
			Image:           pipelineImage,
			ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
			VolumeMounts: []apiv1.VolumeMount{
				{
					Name:      compilerVolumeName,
					MountPath: compilerVolumePath,
				},
			},
			Command: []string{"python3"},
			Args:    args,
		},
		InitContainers: []argo.UserContainer{
			{
				Container: apiv1.Container{
					Name:            CompileStepName,
					Image:           workflows.Config.CompilerImage,
					ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
				},
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

func (workflows *Workflows) uploader(pipelineName string) argo.Template {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp --endpoint %s --output json pipeline upload -p %s %s  | jq -r '.\"Pipeline Details\".\"ID\"'", workflows.Config.KfpEndpoint, pipelineName, PipelineYamlFilePath)

	return argo.Template{
		Name: UploadStepName,
		Inputs: argo.Inputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineYamlParameterName,
					Path: PipelineYamlFilePath,
				},
			},
		},
		Script: &argo.ScriptTemplate{
			Container: apiv1.Container{
				Image:           workflows.Config.KfpToolsImage,
				ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
				Command:         []string{"ash"},
			},
			Source: script,
		},
	}
}

func (workflows *Workflows) deleter() argo.Template {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp --endpoint %s pipeline delete {{inputs.parameters.pipeline-id}}", workflows.Config.KfpEndpoint)

	return argo.Template{
		Name: DeletionStepName,
		Inputs: argo.Inputs{
			Parameters: []argo.Parameter{
				{
					Name: PipelineIdParameterName,
				},
			},
		},
		Script: &argo.ScriptTemplate{
			Container: apiv1.Container{
				Image:           workflows.Config.KfpToolsImage,
				ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
				Command:         []string{"ash"},
			},
			Source: script,
		},
	}
}

func (workflows *Workflows) updater(version string) argo.Template {
	script :=
		"set -e -o pipefail\n" +
			fmt.Sprintf("kfp --endpoint %s pipeline upload-version -v %s -p {{inputs.parameters.pipeline-id}} %s", workflows.Config.KfpEndpoint, version, PipelineYamlFilePath)

	return argo.Template{
		Name: UpdateStepName,
		Inputs: argo.Inputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineYamlParameterName,
					Path: PipelineYamlFilePath,
				},
			},
			Parameters: []argo.Parameter{
				{
					Name: PipelineIdParameterName,
				},
			},
		},
		Script: &argo.ScriptTemplate{
			Container: apiv1.Container{
				Image:           workflows.Config.KfpToolsImage,
				ImagePullPolicy: apiv1.PullPolicy(workflows.Config.ImagePullPolicy),
				Command:         []string{"ash"},
			},
			Source: script,
		},
	}
}
