//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("RunConfiguration controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			providerId := "12345"
			rcHelper := Create(pipelinesv1.RandomRunConfiguration())

			Eventually(rcHelper.ToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
			})).Should(Succeed())

			Eventually(rcHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(rcHelper.ToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(runConfiguration.Status.ProviderId.Provider).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())

			Expect(rcHelper.Update(func(runConfiguration *pipelinesv1.RunConfiguration) {
				runConfiguration.Spec = pipelinesv1.RandomRunConfigurationSpec()
			})).To(Succeed())

			Eventually(rcHelper.ToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Updating))
			})).Should(Succeed())

			Eventually(rcHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(rcHelper.ToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(runConfiguration.Status.ProviderId.Provider).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())

			Expect(rcHelper.Delete()).To(Succeed())

			Eventually(rcHelper.ToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Deleting))
			})).Should(Succeed())

			Eventually(rcHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: ""})
			})).Should(Succeed())

			Eventually(rcHelper.Exists).Should(Not(Succeed()))

			Eventually(rcHelper.EmittedEventsToMatch(func(g Gomega, events []v1.Event) {
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
			runConfiguration := pipelinesv1.RandomRunConfiguration()
			pipelineVersion := apis.RandomString()
			runConfiguration.Spec.Pipeline = pipelinesv1.PipelineIdentifier{Name: apis.RandomString(), Version: pipelineVersion}

			rcHelper := Create(runConfiguration)

			Eventually(rcHelper.ToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(pipelineVersion))
			})).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with no version specified on the RC", func() {
		It("triggers an update of the run configuration", func() {
			pipeline := pipelinesv1.RandomPipeline()

			runConfiguration := pipelinesv1.RandomRunConfiguration()
			runConfiguration.Spec.Pipeline = pipeline.UnversionedIdentifier()
			runConfiguration.Status.ObservedPipelineVersion = pipeline.ComputeVersion()

			pipelineHelper := CreateStable(pipeline)

			rcHelper := CreateStable(runConfiguration)

			pipelineHelper.UpdateStable(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec()
			})

			Eventually(rcHelper.ToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Updating))
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(pipeline.ComputeVersion()))
			})).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with a fixed version specified on the RC", func() {
		It("does not trigger an update of the run configuration", func() {
			pipeline := pipelinesv1.RandomPipeline()
			fixedIdentifier := pipeline.VersionedIdentifier()

			runConfiguration := pipelinesv1.RandomRunConfiguration()

			runConfiguration.Spec.Pipeline = fixedIdentifier
			runConfiguration.Status.ObservedPipelineVersion = pipeline.ComputeVersion()

			pipelineHelper := CreateStable(pipeline)

			rcHelper := CreateStable(runConfiguration)

			pipelineHelper.UpdateStable(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec()
			})

			// To verify the absence of additional RC updates, force another update of the resource.
			// If the update is processed but the pipeline version hasn't changed,
			// given that reconciliation requests are processed in-order, we can conclude that the RC is fixed.
			rcHelper.Update(func(runConfiguration *pipelinesv1.RunConfiguration) {
				runConfiguration.Spec.Schedule = apis.RandomString()
			})

			Eventually(rcHelper.ToMatch(func(g Gomega, runConfiguration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Updating))
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(fixedIdentifier.Version))
			})).Should(Succeed())
		})
	})
})
