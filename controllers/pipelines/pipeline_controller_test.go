package controllers

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/pipeline_workflows"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/test_utils"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("Pipeline controller k8s integration", func() {

	When("Creating, updating and deleting", func() {

		It("transitions through all stages", func() {
			testCtx := test_utils.NewTestContext(k8sClient, ctx)

			Expect(k8sClient.Create(ctx, testCtx.Pipeline)).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Expect(testCtx.UpdateWorkflow(pipeline_workflows.Create, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				pipeline_workflows.SetWorkflowOutput(workflow, pipeline_workflows.PipelineIdParameterName, test_utils.PipelineId)
			})).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())

			Expect(testCtx.UpdatePipeline(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = test_utils.SpecV2
			})).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			Expect(testCtx.UpdateWorkflow(pipeline_workflows.Update, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())

			Expect(testCtx.DeletePipeline()).To(Succeed())

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Expect(testCtx.UpdateWorkflow(pipeline_workflows.Delete, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())

			Eventually(testCtx.PipelineExists).Should(Not(Succeed()))
		})
	})
})
