//go:build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Pipeline controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			providerId := "12345"
			pipelineHelper := Create(pipelinesv1.RandomPipeline(provider.Name))
			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Creating))
				g.Expect(pipeline.Status.ObservedGeneration).To(Equal(pipeline.GetGeneration()))
			})).Should(Succeed())

			Eventually(pipelineHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(pipeline.Status.ProviderId.Provider).To(Equal(pipeline.Spec.Provider))
			})).Should(Succeed())

			Expect(pipelineHelper.Update(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec(provider.Name)
			})).To(Succeed())

			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Updating))
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Updating))
			})).Should(Succeed())

			Eventually(pipelineHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(pipeline.Status.ProviderId.Provider).To(Equal(pipeline.Spec.Provider))
			})).Should(Succeed())

			Expect(pipelineHelper.Delete()).To(Succeed())

			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
				g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Deleting))
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Deleting))
			})).Should(Succeed())

			Eventually(pipelineHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: ""})
			})).Should(Succeed())

			Eventually(pipelineHelper.Exists).Should(Not(Succeed()))

			Eventually(pipelineHelper.EmittedEventsToMatch(func(g Gomega, events []v1.Event) {
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
