//go:build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowutil"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Experiment controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			providerId := "12345"
			anotherProviderId := "67890"
			experimentHelper := Create(
				pipelineshub.RandomExperiment(
					Provider.GetCommonNamespacedName(),
				),
			)

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelineshub.Experiment) {
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Creating))
				g.Expect(experiment.Status.ObservedGeneration).To(Equal(experiment.GetGeneration()))
			})).Should(Succeed())

			Eventually(experimentHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				workflowutil.SetProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelineshub.Experiment) {
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(experiment.Status.Provider.Name).To(Equal(experiment.Spec.Provider))
			})).Should(Succeed())

			Expect(experimentHelper.Update(func(pipeline *pipelineshub.Experiment) {
				pipeline.Spec = pipelineshub.RandomExperimentSpec(
					Provider.GetCommonNamespacedName(),
				)
			})).To(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelineshub.Experiment) {
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Updating))
			})).Should(Succeed())

			Eventually(experimentHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				workflowutil.SetProviderOutput(workflow, providers.Output{Id: anotherProviderId})
			})).Should(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelineshub.Experiment) {
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(experiment.Status.Provider.Name).To(Equal(experiment.Spec.Provider))
			})).Should(Succeed())

			Expect(experimentHelper.Delete()).To(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelineshub.Experiment) {
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Deleting))
			})).Should(Succeed())

			Eventually(experimentHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				workflowutil.SetProviderOutput(workflow, providers.Output{Id: ""})
			})).Should(Succeed())

			Eventually(experimentHelper.Exists).Should(Not(Succeed()))

			Eventually(experimentHelper.EmittedEventsToMatch(func(g Gomega, events []v1.Event) {
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
