//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("RunConfiguration controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			runConfiguration := RandomRunConfiguration()
			runConfiguration.Namespace = "default"

			testCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)

			Expect(k8sClient.Create(ctx, testCtx.RunConfiguration)).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Creating))
				g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
			})).Should(Succeed())

			testCtx.WorkflowSucceeded(WorkflowConstants.CreateOperationLabel)

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(WorkflowConstants.CreateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.UpdateRunConfiguration(func(runConfiguration *pipelinesv1.RunConfiguration) {
				runConfiguration.Spec = RandomRunConfigurationSpec()
			})).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
			})).Should(Succeed())

			testCtx.WorkflowSucceeded(WorkflowConstants.UpdateOperationLabel)

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(WorkflowConstants.UpdateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.DeleteRunConfiguration()).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Deleting))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(WorkflowConstants.DeleteOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
			})).Should(Succeed())

			Eventually(testCtx.RunConfigurationExists).Should(Not(Succeed()))
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

	When("Creating with a fixed pipeline version", func() {
		It("creates a RC with an ObervedPipelineVersion that matches the fixed version", func() {
			runConfiguration := RandomRunConfiguration()
			runConfiguration.Namespace = "default"
			pipelineVersion := "12345-abcde"
			runConfiguration.Spec.Pipeline = "dummy-pipeline:" + pipelineVersion

			testCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)
			Expect(k8sClient.Create(ctx, testCtx.RunConfiguration)).To(Succeed())

			Eventually(testCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(pipelineVersion))
			})).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with no version specified on the RC", func() {
		It("triggers an update of the run configuration", func() {
			pipeline := RandomPipeline()
			pipeline.Namespace = "default"
			pipelineVersion := RandomString()

			runConfiguration := RandomRunConfiguration()
			runConfiguration.Spec.Pipeline = pipeline.Name
			runConfiguration.Namespace = "default"

			runCfgTestCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)
			runCfgTestCtx.RunConfigurationCreatedWithStatus(pipelinesv1.RunConfigurationStatus{
				Status: pipelinesv1.Status{
					Version:              RandomString(),
					KfpId:                RandomString(),
					SynchronizationState: pipelinesv1.Succeeded,
				},
				ObservedPipelineVersion: RandomString(),
			})

			pipelineTestCtx := NewPipelineTestContext(pipeline, k8sClient, ctx)
			pipelineTestCtx.PipelineCreatedWithStatus(
				pipelinesv1.Status{
					Version:              pipelineVersion,
					KfpId:                RandomString(),
					SynchronizationState: pipelinesv1.Succeeded,
				})

			Eventually(runCfgTestCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				// TODO: The RunConfigurationCreatedWithStatus method (called above) sets the SynchronizationState to Updating
				// regardless of whether the pipeline is updated or not - so what's the point of this assertion?
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(pipelineVersion))
			})).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with a fixed version specified on the RC", func() {
		It("does not trigger an update of the run configuration", func() {
			pipeline := RandomPipeline()
			pipeline.Namespace = "default"
			pipelineVersion := RandomString()

			runConfiguration := RandomRunConfiguration()
			fixedPipelineVersion := "12345-abcde"
			runConfiguration.Spec.Pipeline = pipeline.Name + ":" + fixedPipelineVersion
			runConfiguration.Namespace = "default"

			runCfgTestCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)
			Expect(k8sClient.Create(ctx, runCfgTestCtx.RunConfiguration)).To(Succeed())

			pipelineTestCtx := NewPipelineTestContext(pipeline, k8sClient, ctx)
			pipelineTestCtx.PipelineCreatedWithStatus(
				pipelinesv1.Status{
					Version:              pipelineVersion,
					KfpId:                RandomString(),
					SynchronizationState: pipelinesv1.Succeeded,
				})

			Eventually(runCfgTestCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				// TODO - Similar to the TODO comment above: there's no way to check that the RunConfig doesn't enter an Updating state
				// after the pipeline is updated. The following will be Creating.
				g.Expect(runConfiguration.Status.SynchronizationState).NotTo(Equal(pipelinesv1.Updating))
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(fixedPipelineVersion))
			})).Should(Succeed())
		})
	})
})
