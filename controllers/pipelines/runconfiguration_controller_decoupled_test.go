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

var _ = Describe("RunConfiguration controller k8s integration", func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			runConfiguration := RandomRunConfiguration()
			runConfiguration.Namespace = "default"

			kfpId := "12345"
			anotherKfpId := "67890"
			testCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)

			Expect(k8sClient.Create(ctx, testCtx.RunConfiguration)).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(RunConfigurationWorkflowConstants.CreateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutputs(
					workflow,
					[]argo.Parameter{
						{
							Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
							Value: argo.AnyStringPtr(kfpId),
						},
					},
				)
			})).Should(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(RunConfigurationWorkflowConstants.CreateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.UpdateRunConfiguration(func(pipeline *pipelinesv1.RunConfiguration) {
				pipeline.Spec = RandomRunConfigurationSpec()
			})).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(RunConfigurationWorkflowConstants.UpdateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutputs(
					workflow,
					[]argo.Parameter{
						{
							Name:  RunConfigurationWorkflowConstants.RunConfigurationIdParameterName,
							Value: argo.AnyStringPtr(anotherKfpId),
						},
					},
				)
			})).Should(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(RunConfigurationWorkflowConstants.UpdateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.DeleteRunConfiguration()).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(RunConfigurationWorkflowConstants.DeleteOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).Should(Succeed())

			Eventually(testCtx.RunConfigurationExists).Should(Not(Succeed()))
			Eventually(testCtx.FetchWorkflow(RunConfigurationWorkflowConstants.DeleteOperationLabel)).Should(Not(Succeed()))

			Eventually(testCtx.EmittedEventsToMatch(func (g Gomega, events []v1.Event) {
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
