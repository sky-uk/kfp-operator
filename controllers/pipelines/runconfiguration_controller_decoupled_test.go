//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1"
)

var _ = Describe("RunConfiguration controller k8s integration", func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			runConfiguration := RandomRunConfiguration()
			runConfiguration.Namespace = "default"

			kfpId := "12345"
			testCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)

			Expect(k8sClient.Create(ctx, testCtx.RunConfiguration)).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
			})).Should(Succeed())

			Expect(testCtx.UpdateWorkflow(RunConfigurationWorkflowConstants.CreateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setWorkflowOutput(workflow, RunConfigurationWorkflowConstants.RunConfigurationIdParameterName, kfpId)
			})).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())

			Expect(testCtx.UpdateRunConfiguration(func(pipeline *pipelinesv1.RunConfiguration) {
				pipeline.Spec = RandomRunConfigurationSpec()
			})).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			Expect(testCtx.UpdateWorkflow(RunConfigurationWorkflowConstants.UpdateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())

			Expect(testCtx.DeleteRunConfiguration()).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, pipeline *pipelinesv1.RunConfiguration) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Expect(testCtx.UpdateWorkflow(RunConfigurationWorkflowConstants.DeleteOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).To(Succeed())

			Eventually(testCtx.RunConfigurationExists).Should(Not(Succeed()))
		})
	})
})
