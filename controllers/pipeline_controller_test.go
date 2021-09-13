package controllers

import (
	"fmt"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/api/v1"
	testing "github.com/sky-uk/kfp-operator/controllers/testing"
	pipelineworkflows "github.com/sky-uk/kfp-operator/controllers/workflows"
	// +kubebuilder:scaffold:imports
)

var _ = Describe("Pipeline controller k8s integration", func() {

	When("Creating, updating and deleting", func() {

		It("transitions through all stages", func() {
			testCtx := testing.NewTestContext(k8sClient, ctx)

			fmt.Println("1")
			Expect(k8sClient.Create(ctx, testCtx.Pipeline)).To(Succeed())
			fmt.Println("2")
			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())
			fmt.Println("3")
			Expect(testCtx.UpdateWorkflow(pipelineworkflows.Create, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				pipelineworkflows.SetWorkflowOutput(workflow, pipelineworkflows.PipelineIdParameterName, testing.PipelineId)
			})).To(Succeed())
			fmt.Println("4")

			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			fmt.Println("5")
			Expect(testCtx.UpdatePipeline(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = testing.SpecV2
			})).To(Succeed())
			fmt.Println("6")
			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())
			fmt.Println("7")
			Expect(testCtx.UpdateWorkflow(pipelineworkflows.Update, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())
			fmt.Println("8")
			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			fmt.Println("9")
			Expect(testCtx.DeletePipeline()).To(Succeed())
			fmt.Println("10")
			Eventually(testCtx.PipelineToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())
			fmt.Println("12")
			Expect(testCtx.UpdateWorkflow(pipelineworkflows.Delete, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())
			fmt.Println("13")
			Eventually(testCtx.PipelineExists).Should(Not(Succeed()))
			fmt.Println("14")
		})
	})
})
