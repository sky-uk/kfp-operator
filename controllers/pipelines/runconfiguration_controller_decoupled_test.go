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

var _ = Describe("RunConfiguration controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			pipeline := RandomPipeline()
			pipeline.Namespace = "default"

			pipelineTestCtx := NewPipelineTestContext(pipeline, k8sClient, ctx)
			pipelineTestCtx.PipelineCreatedWithStatus(
				pipelinesv1.Status{
					Version: pipeline.Spec.ComputeVersion(),
					KfpId: RandomString(),
					SynchronizationState: pipelinesv1.Succeeded,
				})

			runConfiguration := RandomRunConfiguration()
			runConfiguration.Namespace = "default"
			runConfiguration.Spec.PipelineName = pipeline.Name

			testCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)

			Expect(k8sClient.Create(ctx, testCtx.RunConfiguration)).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
				g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
			})).Should(Succeed())

			testCtx.WorkflowSucceeded(RunConfigurationWorkflowConstants.CreateOperationLabel)

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(RunConfigurationWorkflowConstants.CreateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.UpdateRunConfiguration(func(runConfiguration *pipelinesv1.RunConfiguration) {
				runConfiguration.Spec = RandomRunConfigurationSpec()
				runConfiguration.Spec.PipelineName = pipeline.Name
			})).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			testCtx.WorkflowSucceeded(RunConfigurationWorkflowConstants.UpdateOperationLabel)

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(RunConfigurationWorkflowConstants.UpdateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.DeleteRunConfiguration()).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(RunConfigurationWorkflowConstants.DeleteOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).Should(Succeed())

			Eventually(testCtx.RunConfigurationExists).Should(Not(Succeed()))
			Eventually(testCtx.FetchWorkflow(RunConfigurationWorkflowConstants.DeleteOperationLabel)).Should(Not(Succeed()))

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

	When("The referenced pipeline is missing", func() {
		It("fails the run configuration", func() {
			runConfiguration := RandomRunConfiguration()
			runConfiguration.Namespace = "default"

			testCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)

			Expect(k8sClient.Create(ctx, testCtx.RunConfiguration)).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Failed))
			})).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline", func() {
		It("updates the run configuration", func() {
			pipeline := RandomPipeline()
			pipeline.Namespace = "default"

			pipelineTestCtx := NewPipelineTestContext(pipeline, k8sClient, ctx)
			pipelineTestCtx.PipelineCreatedWithStatus(
				pipelinesv1.Status{
					Version: RandomString(),
					KfpId: RandomString(),
					SynchronizationState: pipelinesv1.Succeeded,
				})

			runConfiguration := RandomRunConfiguration()
			runConfiguration.Spec.PipelineName = pipeline.Name
			runConfiguration.Namespace = "default"

			runCfgTestCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)

			runCfgTestCtx.RunConfigurationCreatedWithStatus(pipelinesv1.RunConfigurationStatus{
				Status: pipelinesv1.Status{
					Version: RandomString(),
					KfpId: RandomString(),
					SynchronizationState: pipelinesv1.Succeeded,
				},
				DesiredPipelineVersion: RandomString(),
			})

			Expect(pipelineTestCtx.UpdatePipelineStatus(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Status.Version = RandomString()
			})).To(Succeed())

			Eventually(runCfgTestCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			runCfgTestCtx.WorkflowSucceeded(RunConfigurationWorkflowConstants.UpdateOperationLabel)

			Eventually(runCfgTestCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
		})
	})
})
