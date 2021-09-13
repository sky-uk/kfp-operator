package workflows

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var wf = Workflows{
	Config: Configuration{
		CompilerImage: "image:v1",
		KfpToolsImage: "image:v1",
	},
}

var _ = Describe("Creation Workflow", func() {
	When("creating a workflow with valid paramters", func() {
		It("creates a valid workflow", func() {
			pipeline := RandomPipeline()
			workflow, error := wf.ConstructCreationWorkflow(pipeline)
			Expect(error).NotTo(HaveOccurred())

			Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(OperationLabelKey, Create))
			Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(PipelineLabelKey, pipeline.Name))
		})
	})
})

var _ = Describe("Update Workflow", func() {
	When("creating a workflow with valid paramters", func() {
		It("creates a valid workflow", func() {
			pipeline := RandomPipeline()
			workflow, error := wf.ConstructUpdateWorkflow(pipeline)
			Expect(error).NotTo(HaveOccurred())

			Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(OperationLabelKey, Update))
			Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(PipelineLabelKey, pipeline.Name))
		})
	})
})

var _ = Describe("Deletion Workflow", func() {
	When("creating a workflow with valid paramters", func() {
		It("creates a valid workflow", func() {
			pipeline := RandomPipeline()
			workflow := wf.ConstructDeletionWorkflow(pipeline)

			Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(OperationLabelKey, Delete))
			Expect(workflow.ObjectMeta.Labels).To(HaveKeyWithValue(PipelineLabelKey, pipeline.Name))
		})
	})
})
