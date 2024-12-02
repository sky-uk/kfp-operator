//go:build decoupled

package experiment

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	common "github.com/sky-uk/kfp-operator/controllers/pipelines"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Experiment controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			providerId := "12345"
			anotherProviderId := "67890"
			experimentHelper := common.Create(pipelinesv1.RandomExperiment(common.Provider.Name))

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Creating))
				g.Expect(experiment.Status.ObservedGeneration).To(Equal(experiment.GetGeneration()))
			})).Should(Succeed())

			Eventually(experimentHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				common.SetProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(experiment.Status.Provider.Name).To(Equal(experiment.Spec.Provider))
			})).Should(Succeed())

			Expect(experimentHelper.Update(func(pipeline *pipelinesv1.Experiment) {
				pipeline.Spec = pipelinesv1.RandomExperimentSpec(common.Provider.Name)
			})).To(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Updating))
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Updating))
			})).Should(Succeed())

			Eventually(experimentHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				common.SetProviderOutput(workflow, providers.Output{Id: anotherProviderId})
			})).Should(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(experiment.Status.Provider.Name).To(Equal(experiment.Spec.Provider))
			})).Should(Succeed())

			Expect(experimentHelper.Delete()).To(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Deleting))
				g.Expect(experiment.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Deleting))
			})).Should(Succeed())

			Eventually(experimentHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				common.SetProviderOutput(workflow, providers.Output{Id: ""})
			})).Should(Succeed())

			Eventually(experimentHelper.Exists).Should(Not(Succeed()))

			Eventually(experimentHelper.EmittedEventsToMatch(func(g Gomega, events []v1.Event) {
				g.Expect(events).To(ConsistOf(
					common.HaveReason(common.EventReasons.Syncing),
					common.HaveReason(common.EventReasons.Synced),
					common.HaveReason(common.EventReasons.Syncing),
					common.HaveReason(common.EventReasons.Synced),
					common.HaveReason(common.EventReasons.Syncing),
					common.HaveReason(common.EventReasons.Synced),
				))
			})).Should(Succeed())
		})
	})
})
