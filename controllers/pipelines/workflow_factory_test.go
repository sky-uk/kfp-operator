//go:build unit
// +build unit

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	configv1 "github.com/sky-uk/kfp-operator/apis/config/v1"
	"gopkg.in/yaml.v2"
)

var wf = WorkflowFactory{
	Config: configv1.Configuration{
		DataflowProject: "project",
		PipelineStorage: "gs://some-bucket",
		CompilerImage:   "image:v1",
		KfpToolsImage:   "image:v1",
	},
}

func workflowTemplates(wf argo.Workflow) map[string]argo.Template {
	result := make(map[string]argo.Template, len(wf.Spec.Templates))

	for _, template := range wf.Spec.Templates {
		result[template.Name] = template
	}

	return result
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

var _ = Describe("Pipeline Yaml", func() {
	When("Called with valid input", func() {
		It("Creates a valid YAML", func() {

			pipeline := RandomPipeline()
			createdYml, err := wf.pipelineConfigAsYaml(pipeline.Spec, pipeline.ObjectMeta)
			Expect(err).NotTo(HaveOccurred())
			m := make(map[interface{}]interface{})
			yaml.Unmarshal([]byte(createdYml), m)

			pipelineRoot := "gs://some-bucket/" + pipeline.Name

			Expect(m["name"]).To(Equal(pipeline.Name))
			Expect(m["pipelineRoot"]).To(Equal(pipelineRoot))
			Expect(m["servingDir"]).To(Equal(pipelineRoot + "/serving"))
			beamArgs := m["beamArgs"].(map[interface{}]interface{})
			Expect(beamArgs["project"]).To(Equal(wf.Config.DataflowProject))
			Expect(beamArgs["temp_location"]).To(Equal(pipelineRoot + "/tmp"))
			spec := m["spec"].(map[interface{}]interface{})
			Expect(spec["image"]).To(Equal(pipeline.Spec.Image))
			Expect(spec["tfxComponents"]).To(Equal(pipeline.Spec.TfxComponents))
			env := spec["env"].(map[interface{}]interface{})
			for k, v := range pipeline.Spec.Env {
				Expect(env[k]).To(Equal(v))
			}
		})
	})
})
