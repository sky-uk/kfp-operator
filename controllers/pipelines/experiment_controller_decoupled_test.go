//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Experiment controller k8s integration", func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			experiment := RandomExperiment()
			experiment.Namespace = "default"

			kfpId := "12345"
			anotherKfpId := "67890"
			testCtx := NewExperimentTestContext(experiment, k8sClient, ctx)

			Expect(k8sClient.Create(ctx, testCtx.Experiment)).To(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, pipeline *pipelinesv1.Experiment) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(ExperimentWorkflowConstants.CreateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutputs(
					workflow,
					[]argo.Parameter{
						{
							Name:  ExperimentWorkflowConstants.ExperimentIdParameterName,
							Value: argo.AnyStringPtr(kfpId),
						},
					},
				)
			})).Should(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, pipeline *pipelinesv1.Experiment) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(ExperimentWorkflowConstants.CreateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.UpdateExperiment(func(pipeline *pipelinesv1.Experiment) {
				pipeline.Spec = RandomExperimentSpec()
			})).To(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, pipeline *pipelinesv1.Experiment) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(ExperimentWorkflowConstants.UpdateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutputs(
					workflow,
					[]argo.Parameter{
						{
							Name:  ExperimentWorkflowConstants.ExperimentIdParameterName,
							Value: argo.AnyStringPtr(anotherKfpId),
						},
					},
				)
			})).Should(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, pipeline *pipelinesv1.Experiment) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(ExperimentWorkflowConstants.UpdateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.DeleteExperiment()).To(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, pipeline *pipelinesv1.Experiment) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(ExperimentWorkflowConstants.DeleteOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).Should(Succeed())

			Eventually(testCtx.ExperimentExists).Should(Not(Succeed()))
			Eventually(testCtx.FetchWorkflow(ExperimentWorkflowConstants.DeleteOperationLabel)).Should(Not(Succeed()))

			Eventually(testCtx.EmittedEventsToMatch(func(g Gomega, events []v1.Event) {
				g.Expect(events).To(ConsistOf(
					HaveReason(string(pipelinesv1.Creating)),
					HaveReason(string(pipelinesv1.Succeeded)),
					HaveReason(string(pipelinesv1.Updating)),
					HaveReason(string(pipelinesv1.Succeeded)),
					HaveReason(string(pipelinesv1.Deleting)),
					HaveReason(string(pipelinesv1.Deleted)),
				))
			})).Should(Succeed())
		})
	})
})
