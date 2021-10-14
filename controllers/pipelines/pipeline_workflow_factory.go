package pipelines

import (
	"fmt"
	"gopkg.in/yaml.v2"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var PipelineWorkflowConstants = struct {
	PipelineIdParameterName   string
	PipelineYamlParameterName string
	CompileStepName           string
	UploadStepName            string
	DeletionStepName          string
	UpdateStepName            string
	PipelineYamlFilePath      string
	PipelineIdFilePath        string
	PipelineNameLabelKey      string
	OperationLabelKey         string
	CreateOperationLabel      string
	UpdateOperationLabel      string
	DeleteOperationLabel      string
}{
	PipelineIdParameterName:   "pipeline-id",
	PipelineYamlParameterName: "pipeline",
	CompileStepName:           "compile",
	UploadStepName:            "upload",
	DeletionStepName:          "delete",
	UpdateStepName:            "update",
	PipelineYamlFilePath:      "/tmp/pipeline.yaml",
	PipelineIdFilePath:        "/tmp/pipeline.txt",
	PipelineNameLabelKey:      "pipelines.kubeflow.org/pipeline",
	OperationLabelKey:         "pipelines.kubeflow.org/operation",
	CreateOperationLabel:      "create-pipeline",
	UpdateOperationLabel:      "update-pipeline",
	DeleteOperationLabel:      "delete-pipeline",
}

var (
	// Needs to be passed by reference
	trueValue = true
)

type PipelineWorkflowFactory struct {
	WorkflowFactory
}

type CompilerConfig struct {
	RootLocation    string            `yaml:"rootLocation"`
	ServingLocation string            `yaml:"servingLocation"`
	Name            string            `yaml:"name"`
	Image           string            `yaml:"image"`
	TfxComponents   string            `yaml:"tfxComponents"`
	Env             map[string]string `yaml:"env"`
	BeamArgs        map[string]string `yaml:"beamArgs"`
}

func (config CompilerConfig) AsYaml() (string, error) {
	configYaml, err := yaml.Marshal(&config)

	if err != nil {
		return "", err
	}

	return string(configYaml), nil
}

// TODO: Join paths properly (path.Join or filepath.Join don't work with URLs)
func (wf *PipelineWorkflowFactory) newCompilerConfig(pipelineSpec pipelinesv1.PipelineSpec, pipelineMeta metav1.ObjectMeta) *CompilerConfig {
	// TODO: should come from config
	servingPath := "/serving"
	tempPath := "/tmp"

	pipelineRoot := wf.Config.PipelineStorage + "/" + pipelineMeta.Name

	beamArgs := make(map[string]string)
	for key, value := range wf.Config.DefaultBeamArgs {
		beamArgs[key] = value
	}
	for key, value := range pipelineSpec.BeamArgs {
		beamArgs[key] = value
	}
	beamArgs["temp_location"] = pipelineRoot + tempPath

	return &CompilerConfig{
		RootLocation:    pipelineRoot,
		ServingLocation: pipelineRoot + servingPath,
		Name:            pipelineMeta.Name,
		Image:           pipelineSpec.Image,
		TfxComponents:   pipelineSpec.TfxComponents,
		Env:             pipelineSpec.Env,
		BeamArgs:        beamArgs,
	}
}

func (w *PipelineWorkflowFactory) commonMeta(pipelineMeta metav1.ObjectMeta, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-",
		Namespace:    pipelineMeta.Namespace,
		Labels: map[string]string{
			PipelineWorkflowConstants.OperationLabelKey:    operation,
			PipelineWorkflowConstants.PipelineNameLabelKey: pipelineMeta.Name,
		},
	}
}

func (w PipelineWorkflowFactory) ConstructCreationWorkflow(pipelineSpec pipelinesv1.PipelineSpec, pipelineMeta metav1.ObjectMeta, pipelineVersion string) (*argo.Workflow, error) {
	compilerConfigYaml, error := w.newCompilerConfig(pipelineSpec, pipelineMeta).AsYaml()

	if error != nil {
		return nil, error
	}

	entrypointName := PipelineWorkflowConstants.CreateOperationLabel

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(pipelineMeta, PipelineWorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     PipelineWorkflowConstants.CompileStepName,
									Template: PipelineWorkflowConstants.CompileStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     PipelineWorkflowConstants.UploadStepName,
									Template: PipelineWorkflowConstants.UploadStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: PipelineWorkflowConstants.PipelineYamlParameterName,
												From: fmt.Sprintf("{{steps.%s.outputs.artifacts.pipeline}}",
													PipelineWorkflowConstants.CompileStepName),
											},
										},
									},
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     PipelineWorkflowConstants.UpdateStepName,
									Template: PipelineWorkflowConstants.UpdateStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: PipelineWorkflowConstants.PipelineYamlParameterName,
												From: fmt.Sprintf("{{steps.%s.outputs.artifacts.pipeline}}",
													PipelineWorkflowConstants.CompileStepName),
											},
										},
										Parameters: []argo.Parameter{
											{
												Name: PipelineWorkflowConstants.PipelineIdParameterName,
												Value: argo.AnyStringPtr(fmt.Sprintf("{{steps.%s.outputs.result}}",
													PipelineWorkflowConstants.UploadStepName)),
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
								Name: PipelineWorkflowConstants.PipelineIdParameterName,
								ValueFrom: &argo.ValueFrom{
									Parameter: fmt.Sprintf("{{steps.%s.outputs.result}}",
										PipelineWorkflowConstants.UploadStepName),
								},
							},
						},
					},
				},
				w.compiler(compilerConfigYaml, pipelineSpec.Image),
				w.uploader(pipelineMeta.Name),
				w.updater(pipelineVersion),
			},
		},
	}

	return workflow, nil
}

func (w PipelineWorkflowFactory) ConstructUpdateWorkflow(pipelineSpec pipelinesv1.PipelineSpec, pipelineMeta metav1.ObjectMeta, pipelineId string, pipelineVersion string) (*argo.Workflow, error) {
	compilerConfigYaml, error := w.newCompilerConfig(pipelineSpec, pipelineMeta).AsYaml()

	if error != nil {
		return nil, error
	}

	entrypointName := PipelineWorkflowConstants.UpdateOperationLabel

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(pipelineMeta, PipelineWorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     PipelineWorkflowConstants.CompileStepName,
									Template: PipelineWorkflowConstants.CompileStepName,
								},
							},
						},
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     PipelineWorkflowConstants.UpdateStepName,
									Template: PipelineWorkflowConstants.UpdateStepName,
									Arguments: argo.Arguments{
										Artifacts: []argo.Artifact{
											{
												Name: PipelineWorkflowConstants.PipelineYamlParameterName,
												From: fmt.Sprintf("{{steps.%s.outputs.artifacts.pipeline}}",
													PipelineWorkflowConstants.CompileStepName),
											},
										},
										Parameters: []argo.Parameter{
											{
												Name:  PipelineWorkflowConstants.PipelineIdParameterName,
												Value: argo.AnyStringPtr(pipelineId),
											},
										},
									},
								},
							},
						},
					},
				},
				w.compiler(compilerConfigYaml, pipelineSpec.Image),
				w.updater(pipelineVersion),
			},
		},
	}

	return workflow, nil
}

func (w PipelineWorkflowFactory) ConstructDeletionWorkflow(pipelineMeta metav1.ObjectMeta, pipelineId string) *argo.Workflow {

	entrypointName := PipelineWorkflowConstants.DeleteOperationLabel

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(pipelineMeta, PipelineWorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.ServiceAccount,
			Entrypoint:         entrypointName,
			Templates: []argo.Template{
				{
					Name: entrypointName,
					Steps: []argo.ParallelSteps{
						{
							Steps: []argo.WorkflowStep{
								{
									Name:     PipelineWorkflowConstants.DeletionStepName,
									Template: PipelineWorkflowConstants.DeletionStepName,
									Arguments: argo.Arguments{
										Parameters: []argo.Parameter{
											{
												Name:  PipelineWorkflowConstants.PipelineIdParameterName,
												Value: argo.AnyStringPtr(pipelineId),
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

func (workflows *PipelineWorkflowFactory) compiler(compilerConfigYaml string, pipelineImage string) argo.Template {
	compilerVolumeName := "compiler"
	compilerVolumePath := "/compiler"

	args := []string{
		"/compiler/compiler.py",
		"--output_file",
		PipelineWorkflowConstants.PipelineYamlFilePath,
		"--pipeline_config",
		compilerConfigYaml,
	}

	return argo.Template{
		Name: PipelineWorkflowConstants.CompileStepName,
		Outputs: argo.Outputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineWorkflowConstants.PipelineYamlParameterName,
					Path: PipelineWorkflowConstants.PipelineYamlFilePath,
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
					Name:            PipelineWorkflowConstants.CompileStepName,
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

func (workflows *PipelineWorkflowFactory) uploader(pipelineName string) argo.Template {
	kfpScript := fmt.Sprintf(`pipeline upload --pipeline-name %s %s  | jq -r '."Pipeline Details"."ID"'`,
		pipelineName, PipelineWorkflowConstants.PipelineYamlFilePath)

	return argo.Template{
		Name: PipelineWorkflowConstants.UploadStepName,
		Inputs: argo.Inputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineWorkflowConstants.PipelineYamlParameterName,
					Path: PipelineWorkflowConstants.PipelineYamlFilePath,
				},
			},
		},
		Script: workflows.ScriptTemplate(kfpScript),
	}
}

func (workflows *PipelineWorkflowFactory) deleter() argo.Template {
	kfpScript := "pipeline delete {{inputs.parameters.pipeline-id}}"

	return argo.Template{
		Name: PipelineWorkflowConstants.DeletionStepName,
		Inputs: argo.Inputs{
			Parameters: []argo.Parameter{
				{
					Name: PipelineWorkflowConstants.PipelineIdParameterName,
				},
			},
		},
		Script: workflows.ScriptTemplate(kfpScript),
	}
}

func (workflows *PipelineWorkflowFactory) updater(version string) argo.Template {
	kfpScript := fmt.Sprintf("pipeline upload-version --pipeline-version %s --pipeline-id {{inputs.parameters.pipeline-id}} %s",
		version, PipelineWorkflowConstants.PipelineYamlFilePath)

	return argo.Template{
		Name: PipelineWorkflowConstants.UpdateStepName,
		Inputs: argo.Inputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineWorkflowConstants.PipelineYamlParameterName,
					Path: PipelineWorkflowConstants.PipelineYamlFilePath,
				},
			},
			Parameters: []argo.Parameter{
				{
					Name: PipelineWorkflowConstants.PipelineIdParameterName,
				},
			},
		},
		Script: workflows.ScriptTemplate(kfpScript),
	}
}
