//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("RunConfiguration controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			runConfiguration := RandomRunConfiguration()

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
		It("creates a RC with an ObservedPipelineVersion that matches the fixed version", func() {
			runConfiguration := RandomRunConfiguration()
			pipelineVersion := RandomString()
			runConfiguration.Spec.Pipeline = pipelinesv1.PipelineIdentifier{Name: RandomString(), Version: pipelineVersion}

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

			runConfiguration := RandomRunConfiguration()
			runConfiguration.Spec.Pipeline = pipeline.UnversionedIdentifier()
			runConfiguration.Status.ObservedPipelineVersion = pipeline.Spec.ComputeVersion()

			pipelineTestCtx := NewPipelineTestContext(pipeline, k8sClient, ctx)
			pipelineTestCtx.StablePipelineCreated()

			runCfgTestCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)
			runCfgTestCtx.StableRunConfigurationCreated()

			newPipelineSpec := RandomPipelineSpec()
			pipelineTestCtx.StablePipelineUpdated(newPipelineSpec)

			Eventually(runCfgTestCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(newPipelineSpec.ComputeVersion()))
			})).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with a fixed version specified on the RC", func() {
		It("does not trigger an update of the run configuration", func() {
			pipeline := RandomPipeline()
			fixedIdentifier := pipeline.VersionedIdentifier()

			runConfiguration := RandomRunConfiguration()
			runConfiguration.Spec.Pipeline = fixedIdentifier
			runConfiguration.Status.ObservedPipelineVersion = pipeline.Spec.ComputeVersion()

			pipelineTestCtx := NewPipelineTestContext(pipeline, k8sClient, ctx)
			pipelineTestCtx.StablePipelineCreated()

			runCfgTestCtx := NewRunConfigurationTestContext(runConfiguration, k8sClient, ctx)
			runCfgTestCtx.StableRunConfigurationCreated()

			newPipelineSpec := RandomPipelineSpec()
			pipelineTestCtx.StablePipelineUpdated(newPipelineSpec)

			// To verify the absence of additional RC updates, force another update of the resource.
			// If the update is processed but the pipeline version hasn't changed,
			// given that reconciliation requests are processed in-order, we can conclude that the RC is fixed.
			runCfgTestCtx.UpdateRunConfiguration(func(runConfiguration *pipelinesv1.RunConfiguration) {
				runConfiguration.Spec.Schedule = RandomString()
			})

			Eventually(runCfgTestCtx.RunConfigurationToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(pipelinesv1.Updating))
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(fixedIdentifier.Version))
			})).Should(Succeed())
		})
	})
})
