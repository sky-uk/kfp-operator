//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Experiment controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			providerId := "12345"
			anotherProviderId := "67890"
			experimentHelper := Create(pipelinesv1.RandomExperiment())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(experiment.Status.ObservedGeneration).To(Equal(experiment.GetGeneration()))
			})).Should(Succeed())

			Eventually(experimentHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(experiment.Status.ProviderId.Provider).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())

			Expect(experimentHelper.Update(func(pipeline *pipelinesv1.Experiment) {
				pipeline.Spec = pipelinesv1.RandomExperimentSpec()
			})).To(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Updating))
			})).Should(Succeed())

			Eventually(experimentHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: anotherProviderId})
			})).Should(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(experiment.Status.ProviderId.Provider).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())

			Expect(experimentHelper.Delete()).To(Succeed())

			Eventually(experimentHelper.ToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Deleting))
			})).Should(Succeed())

			Eventually(experimentHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: ""})
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
