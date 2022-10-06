package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	providers "github.com/sky-uk/kfp-operator/providers/base"
	"gopkg.in/yaml.v2"
)

var PipelineWorkflowConstants = struct {
	PipelineIdParameterName      string
	PipelineVersionParameterName string
	PipelineConfigParameterName  string
}{
	PipelineIdParameterName:      "pipeline-id",
	PipelineConfigParameterName:  "pipeline-config",
	PipelineVersionParameterName: "pipeline-version",
}

type PipelineWorkflowFactory struct {
	WorkflowFactoryBase
}

func NamedValuesToMap(namedValues []apis.NamedValue) map[string]string {
	m := make(map[string]string)

	for _, nv := range namedValues {
		m[nv.Name] = nv.Value
	}

	return m
}

func NamedValuesToMultiMap(namedValues []apis.NamedValue) map[string][]string {
	multimap := make(map[string][]string)

	for _, nv := range namedValues {
		if _, found := multimap[nv.Name]; !found {
			multimap[nv.Name] = []string{}
		}

		multimap[nv.Name] = append(multimap[nv.Name], nv.Value)
	}

	return multimap
}

// TODO: Join paths properly (path.Join or filepath.Join don't work with URLs)
func (wf *PipelineWorkflowFactory) newCompilerConfig(pipeline *pipelinesv1.Pipeline) *providers.PipelineConfig {
	// TODO: should come from config
	servingPath := "/serving"
	tempPath := "/tmp"

	pipelineRoot := wf.Config.PipelineStorage + "/" + pipeline.ObjectMeta.Name

	beamArgs := append(wf.Config.DefaultBeamArgs, pipeline.Spec.BeamArgs...)
	beamArgs = append(beamArgs, apis.NamedValue{Name: "temp_location", Value: pipelineRoot + tempPath})

	return &providers.PipelineConfig{
		RootLocation:    pipelineRoot,
		ServingLocation: pipelineRoot + servingPath,
		Name:            pipeline.ObjectMeta.Name,
		Version:         pipeline.Spec.ComputeVersion(),
		Image:           pipeline.Spec.Image,
		TfxComponents:   pipeline.Spec.TfxComponents,
		Env:             NamedValuesToMap(pipeline.Spec.Env),
		BeamArgs:        NamedValuesToMultiMap(beamArgs),
	}
}

func (workflows PipelineWorkflowFactory) ConstructCreationWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	compilerConfigYaml, err := yaml.Marshal(workflows.newCompilerConfig(pipeline))
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(pipeline, WorkflowConstants.CreateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  PipelineWorkflowConstants.PipelineConfigParameterName,
						Value: argo.AnyStringPtr(string(compilerConfigYaml)),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "create-pipeline",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows PipelineWorkflowFactory) ConstructUpdateWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	compilerConfigYaml, err := yaml.Marshal(workflows.newCompilerConfig(pipeline))
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(pipeline, WorkflowConstants.UpdateOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  PipelineWorkflowConstants.PipelineConfigParameterName,
						Value: argo.AnyStringPtr(string(compilerConfigYaml)),
					},
					{
						Name:  PipelineWorkflowConstants.PipelineIdParameterName,
						Value: argo.AnyStringPtr(pipeline.Status.KfpId),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "update-pipeline",
				ClusterScope: true,
			},
		},
	}, nil
}

func (workflows PipelineWorkflowFactory) ConstructDeletionWorkflow(pipeline *pipelinesv1.Pipeline) (*argo.Workflow, error) {
	compilerConfigYaml, err := yaml.Marshal(workflows.newCompilerConfig(pipeline))
	if err != nil {
		return nil, err
	}

	return &argo.Workflow{
		ObjectMeta: *CommonWorkflowMeta(pipeline, WorkflowConstants.DeleteOperationLabel),
		Spec: argo.WorkflowSpec{
			Arguments: argo.Arguments{
				Parameters: []argo.Parameter{
					{
						Name:  PipelineWorkflowConstants.PipelineIdParameterName,
						Value: argo.AnyStringPtr(pipeline.Status.KfpId),
					},
					{
						Name:  WorkflowConstants.ProviderConfigParameterName,
						Value: argo.AnyStringPtr(workflows.ProviderConfig),
					},
					{
						Name:  PipelineWorkflowConstants.PipelineConfigParameterName,
						Value: argo.AnyStringPtr(string(compilerConfigYaml)),
					},
				},
			},
			WorkflowTemplateRef: &argo.WorkflowTemplateRef{
				Name:         workflows.Config.WorkflowTemplatePrefix + "delete-pipeline",
				ClusterScope: true,
			},
		},
	}, nil
}
