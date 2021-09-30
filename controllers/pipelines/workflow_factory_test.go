//go:build unit
// +build unit

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func workflowTemplates(wf argo.Workflow) map[string]argo.Template {
	result := make(map[string]argo.Template, len(wf.Spec.Templates))

	for _, template := range wf.Spec.Templates {
		result[template.Name] = template
	}

	return result
}

var _ = Describe("Workflows", func() {

	var wf = WorkflowFactory{
		Config: configv1.Configuration{
			DataflowProject: "project",
			PipelineStorage: "gs://some-bucket",
			CompilerImage:   "image:v1",
			KfpToolsImage:   "image:v1",
		},
	}

	var _ = Describe("Creation Workflow", func() {
		When("creating a workflow with valid paramters", func() {
			It("creates a valid workflow", func() {
				pipeline := RandomPipeline()
				workflow, error := wf.ConstructCreationWorkflow(pipeline.Spec, pipeline.ObjectMeta, pipeline.Status.Version)
				Expect(error).NotTo(HaveOccurred())

				Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(OperationLabelKey, CreateOperationLabel))
				Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(PipelineNameLabelKey, pipeline.Name))
				Expect(workflow.ObjectMeta.Namespace).To(Equal(pipeline.Namespace))
				Expect(workflow.Spec.ServiceAccountName).To(Equal(wf.Config.ServiceAccount))
				Expect(workflowTemplates(*workflow)).To(HaveKey(workflow.Spec.Entrypoint))
			})
		})
	})

	var _ = Describe("Update Workflow", func() {
		When("creating a workflow with valid paramters", func() {
			It("creates a valid workflow", func() {
				pipeline := RandomPipeline()
				workflow, error := wf.ConstructUpdateWorkflow(pipeline.Spec, pipeline.ObjectMeta, pipeline.Status.Id, pipeline.Status.Version)
				Expect(error).NotTo(HaveOccurred())

				Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(OperationLabelKey, UpdateOperationLabel))
				Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(PipelineNameLabelKey, pipeline.Name))
				Expect(workflow.ObjectMeta.Namespace).To(Equal(pipeline.Namespace))
				Expect(workflow.Spec.ServiceAccountName).To(Equal(wf.Config.ServiceAccount))
			})
		})
	})

	var _ = Describe("Deletion Workflow", func() {
		When("creating a workflow with valid paramters", func() {
			It("creates a valid workflow", func() {
				pipeline := RandomPipeline()
				workflow := wf.ConstructDeletionWorkflow(pipeline.ObjectMeta, pipeline.Status.Id)

				Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(OperationLabelKey, DeleteOperationLabel))
				Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(PipelineNameLabelKey, pipeline.Name))
				Expect(workflow.ObjectMeta.Namespace).To(Equal(pipeline.Namespace))
				Expect(workflow.Spec.ServiceAccountName).To(Equal(wf.Config.ServiceAccount))
			})
		})
	})
})

var _ = Describe("PipelineConfig", func() {

	Specify("Some fields are copied from Pipeline resource", func() {
		wf := WorkflowFactory{}
		pipeline := pipelinesv1.Pipeline{
			ObjectMeta: v1.ObjectMeta{
				Name: "pipelineName",
			},
			Spec: pipelinesv1.PipelineSpec{
				Image:         "pipelineImage",
				TfxComponents: "pipelineTfxComponents",
				Env: map[string]string{
					"ea": "b",
				},
			},
		}
		compilerConfig := wf.newCompilerConfig(pipeline.Spec, pipeline.ObjectMeta)

		Expect(compilerConfig.Name).To(Equal("pipelineName"))
		Expect(compilerConfig.Image).To(Equal("pipelineImage"))
		Expect(compilerConfig.TfxComponents).To(Equal("pipelineTfxComponents"))
		Expect(compilerConfig.Env["ea"]).To(Equal("b"))
	})

	Specify("Paths are appended to PipelineStorage", func() {
		wf := WorkflowFactory{
			Config: configv1.Configuration{
				PipelineStorage: "gs://bucket",
			},
		}
		pipeline := pipelinesv1.Pipeline{
			ObjectMeta: v1.ObjectMeta{
				Name: "pipelineName",
			},
		}

		compilerConfig := wf.newCompilerConfig(pipeline.Spec, pipeline.ObjectMeta)

		Expect(compilerConfig.RootLocation).To(Equal("gs://bucket/pipelineName"))
		Expect(compilerConfig.ServingLocation).To(Equal("gs://bucket/pipelineName/serving"))
	})

	Specify("Original BeamArgs are copied", func() {
		wf := WorkflowFactory{}
		pipeline := pipelinesv1.Pipeline{
			Spec: pipelinesv1.PipelineSpec{
				BeamArgs: map[string]string{
					"a": "b",
				},
			},
		}

		compilerConfig := wf.newCompilerConfig(pipeline.Spec, pipeline.ObjectMeta)

		Expect(compilerConfig.BeamArgs["a"]).To(Equal("b"))
	})

	Specify("BeamArgs are overridden with temp_location", func() {
		wf := WorkflowFactory{
			Config: configv1.Configuration{
				PipelineStorage: "gs://bucket",
			},
		}
		pipeline := pipelinesv1.Pipeline{
			ObjectMeta: v1.ObjectMeta{
				Name: "pipelineName",
			},
			Spec: pipelinesv1.PipelineSpec{
				BeamArgs: map[string]string{
					"temp_location": "will be overridden",
				},
			},
		}

		compilerConfig := wf.newCompilerConfig(pipeline.Spec, pipeline.ObjectMeta)

		Expect(compilerConfig.BeamArgs["temp_location"]).To(Equal("gs://bucket/pipelineName/tmp"))
	})

	// TODO "BeamArgs default to configuration values"
	Specify("Project in the Spec defaults to configuration value", func() {
		wf := WorkflowFactory{
			Config: configv1.Configuration{
				DataflowProject: "dataflowProject",
			},
		}
		pipeline := pipelinesv1.Pipeline{}

		compilerConfig := wf.newCompilerConfig(pipeline.Spec, pipeline.ObjectMeta)

		Expect(compilerConfig.BeamArgs["project"]).To(Equal("dataflowProject"))
	})

	It("Creates a valid YAML", func() {
		config := CompilerConfig{
			RootLocation:    "pipelineRootLocation",
			ServingLocation: "pipelineServingLocation",
			Name:            "pipelineName",
			Image:           "pipelineImage",
			TfxComponents:   "pipelineTfxComponents",
			Env: map[string]string{
				"ea": "eb",
			},
			BeamArgs: map[string]string{
				"ba": "bb",
			},
		}

		configYaml, err := config.AsYaml()
		Expect(err).NotTo(HaveOccurred())

		m := make(map[interface{}]interface{})
		yaml.Unmarshal([]byte(configYaml), m)

		Expect(m["rootLocation"]).To(Equal("pipelineRootLocation"))
		Expect(m["servingLocation"]).To(Equal("pipelineServingLocation"))
		Expect(m["name"]).To(Equal("pipelineName"))
		Expect(m["image"]).To(Equal("pipelineImage"))
		Expect(m["tfxComponents"]).To(Equal("pipelineTfxComponents"))
		env := m["env"].(map[interface{}]interface{})
		Expect(env["ea"]).To(Equal("eb"))
		beamArgs := m["beamArgs"].(map[interface{}]interface{})
		Expect(beamArgs["ba"]).To(Equal("bb"))
	})
})
