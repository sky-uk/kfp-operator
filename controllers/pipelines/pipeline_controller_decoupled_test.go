//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Pipeline controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			pipeline := RandomPipeline()
			pipeline.Namespace = "default"

			kfpId := "12345"
			testCtx := NewPipelineTestContext(pipeline)

			Expect(testCtx.CreateResource()).To(Succeed())

			Eventually(testCtx.ResourceToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
				g.Expect(pipeline.Status.ObservedGeneration).To(Equal(pipeline.GetGeneration()))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(WorkflowConstants.CreateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutputs(
					workflow,
					[]argo.Parameter{
						{
							Name:  PipelineWorkflowConstants.PipelineIdParameterName,
							Value: argo.AnyStringPtr(kfpId),
						},
						{
							Name:  PipelineWorkflowConstants.PipelineVersionParameterName,
							Value: argo.AnyStringPtr(pipeline.Spec.ComputeVersion()),
						},
					},
				)
			})).Should(Succeed())

			Eventually(testCtx.ResourceToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(WorkflowConstants.CreateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.UpdateResource(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = RandomPipelineSpec()
			})).To(Succeed())

			Eventually(testCtx.ResourceToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(WorkflowConstants.UpdateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).Should(Succeed())

			Eventually(testCtx.ResourceToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(WorkflowConstants.UpdateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.DeleteResource()).To(Succeed())

			Eventually(testCtx.ResourceToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(WorkflowConstants.DeleteOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).Should(Succeed())

			Eventually(testCtx.ResourceExists).Should(Not(Succeed()))
			Eventually(testCtx.FetchWorkflow(WorkflowConstants.DeleteOperationLabel)).Should(Not(Succeed()))

			Eventually(testCtx.EmittedEventsToMatch(func(g Gomega, events []v1.Event) {
				g.Expect(events).To(ConsistOf(
					HaveReason(EventReasons.Syncing),
					HaveReason(EventReasons.Synced),
					HaveReason(EventReasons.Syncing),
					HaveReason(EventReasons.Synced),
					HaveReason(EventReasons.Syncing),
					HaveReason(EventReasons.Synced),
				))
			})).Should(Succeed())
		})
	})
})
