package pipelines

import (
	"context"
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
	PipelineNameLabelKey:      pipelinesv1.GroupVersion.Group + "/pipeline",
	OperationLabelKey:         pipelinesv1.GroupVersion.Group + "/operation",
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
func (wf *PipelineWorkflowFactory) newCompilerConfig(pipeline *pipelinesv1.Pipeline) *CompilerConfig {
	// TODO: should come from config
	servingPath := "/serving"
	tempPath := "/tmp"

	pipelineRoot := wf.Config.PipelineStorage + "/" + pipeline.ObjectMeta.Name

	beamArgs := make(map[string]string)
	for key, value := range wf.Config.DefaultBeamArgs {
		beamArgs[key] = value
	}
	for key, value := range pipeline.Spec.BeamArgs {
		beamArgs[key] = value
	}
	beamArgs["temp_location"] = pipelineRoot + tempPath

	return &CompilerConfig{
		RootLocation:    pipelineRoot,
		ServingLocation: pipelineRoot + servingPath,
		Name:            pipeline.ObjectMeta.Name,
		Image:           pipeline.Spec.Image,
		TfxComponents:   pipeline.Spec.TfxComponents,
		Env:             pipeline.Spec.Env,
		BeamArgs:        beamArgs,
	}
}

func (w *PipelineWorkflowFactory) commonMeta(ctx context.Context, pipeline *pipelinesv1.Pipeline, operation string) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		GenerateName: operation + "-",
		Namespace:    pipeline.ObjectMeta.Namespace,
		Labels: map[string]string{
			PipelineWorkflowConstants.OperationLabelKey:    operation,
			PipelineWorkflowConstants.PipelineNameLabelKey: pipeline.ObjectMeta.Name,
		},
		Annotations: w.Annotations(ctx, pipeline.ObjectMeta),
	}
}

func (w PipelineWorkflowFactory) ConstructCreationWorkflow(ctx context.Context, pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	compilerConfigYaml, error := w.newCompilerConfig(pipeline).AsYaml()

	if error != nil {
		return nil, error
	}

	entrypointName := PipelineWorkflowConstants.CreateOperationLabel

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(ctx, pipeline, PipelineWorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.Argo.ServiceAccount,
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
				w.compiler(compilerConfigYaml, pipeline.Spec.Image),
				w.uploader(pipeline.ObjectMeta.Name),
				w.updater(pipeline.Spec.ComputeVersion()),
			},
		},
	}

	return workflow, nil
}

func (w PipelineWorkflowFactory) ConstructUpdateWorkflow(ctx context.Context, pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	compilerConfigYaml, error := w.newCompilerConfig(pipeline).AsYaml()

	if error != nil {
		return nil, error
	}

	entrypointName := PipelineWorkflowConstants.UpdateOperationLabel

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(ctx, pipeline, PipelineWorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.Argo.ServiceAccount,
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
												Value: argo.AnyStringPtr(pipeline.Status.KfpId),
											},
										},
									},
								},
							},
						},
					},
				},
				w.compiler(compilerConfigYaml, pipeline.Spec.Image),
				w.updater(pipeline.Spec.ComputeVersion()),
			},
		},
	}

	return workflow, nil
}

func (w PipelineWorkflowFactory) ConstructDeletionWorkflow(ctx context.Context, pipeline *pipelinesv1.Pipeline) *argo.Workflow {

	entrypointName := PipelineWorkflowConstants.DeleteOperationLabel

	workflow := &argo.Workflow{
		ObjectMeta: *w.commonMeta(ctx, pipeline, PipelineWorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			ServiceAccountName: w.Config.Argo.ServiceAccount,
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
												Value: argo.AnyStringPtr(pipeline.Status.KfpId),
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

	initContainerSpec := workflows.Config.Argo.ContainerDefaults.DeepCopy()
	initContainerSpec.Name = PipelineWorkflowConstants.CompileStepName
	initContainerSpec.Image = workflows.Config.Argo.CompilerImage

	pipelineContainerSpec := workflows.Config.Argo.ContainerDefaults.DeepCopy()
	pipelineContainerSpec.Name = "pipeline"
	pipelineContainerSpec.Image = pipelineImage
	pipelineContainerSpec.VolumeMounts = []apiv1.VolumeMount{
		{
			Name:      compilerVolumeName,
			MountPath: compilerVolumePath,
		},
	}

	pipelineContainerSpec.Command = []string{"python3"}
	pipelineContainerSpec.Args = []string{
		"/compiler/compiler.py",
		"--output_file",
		PipelineWorkflowConstants.PipelineYamlFilePath,
		"--pipeline_config",
		compilerConfigYaml,
	}

	return argo.Template{
		Name:     PipelineWorkflowConstants.CompileStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
		Outputs: argo.Outputs{
			Artifacts: []argo.Artifact{
				{
					Name: PipelineWorkflowConstants.PipelineYamlParameterName,
					Path: PipelineWorkflowConstants.PipelineYamlFilePath,
				},
			},
		},
		Container: pipelineContainerSpec,
		InitContainers: []argo.UserContainer{
			{
				Container:          *initContainerSpec,
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
		Name:     PipelineWorkflowConstants.UploadStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
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
		Name:     PipelineWorkflowConstants.DeletionStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
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
		Name:     PipelineWorkflowConstants.UpdateStepName,
		Metadata: workflows.Config.Argo.MetadataDefaults,
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
