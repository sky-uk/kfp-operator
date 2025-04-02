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

var _ = Describe("Pipeline controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			providerId := "12345"
			pipeline := pipelineshub.RandomPipeline(Provider.GetCommonNamespacedName())
			pipeline.Spec.Framework.Type = TestFramework
			pipelineHelper := Create(pipeline)
			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelineshub.Pipeline) {
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Creating))
				g.Expect(pipeline.Status.ObservedGeneration).To(Equal(pipeline.GetGeneration()))
			})).Should(Succeed())

			Eventually(pipelineHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				workflowutil.SetProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelineshub.Pipeline) {
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(pipeline.Status.Provider.Name).To(Equal(pipeline.Spec.Provider))
			})).Should(Succeed())

			Expect(pipelineHelper.Update(func(pipeline *pipelineshub.Pipeline) {
				pipeline.Spec = pipelineshub.RandomPipelineSpec(Provider.GetCommonNamespacedName())
				pipeline.Spec.Framework.Type = TestFramework
			})).To(Succeed())

			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelineshub.Pipeline) {
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Updating))
			})).Should(Succeed())

			Eventually(pipelineHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				workflowutil.SetProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelineshub.Pipeline) {
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(pipeline.Status.Provider.Name).To(Equal(pipeline.Spec.Provider))
			})).Should(Succeed())

			Expect(pipelineHelper.Delete()).To(Succeed())

			Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelineshub.Pipeline) {
				g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Deleting))
			})).Should(Succeed())

			Eventually(pipelineHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				workflowutil.SetProviderOutput(workflow, providers.Output{Id: ""})
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
