//go:build decoupled

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Provider controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			providerHelper := Create(pipelinesv1.RandomProvider())
			Eventually(providerHelper.ToMatch(func(g Gomega, provider *pipelinesv1.Provider) {
				g.Expect(provider.Status.SynchronizationState).To(Equal(apis.Creating))
				// TODO: DO WE EVEN NEED THESE THO?
				// g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Creating))
				// g.Expect(pipeline.Status.ObservedGeneration).To(Equal(pipeline.GetGeneration()))
			})).Should(Succeed())

			Eventually(providerHelper.UpdateDeployment(func(deployment *appsv1.Deployment) {
				deployment.Status.Conditions = []appsv1.DeploymentCondition{
					{
						Type:   appsv1.DeploymentAvailable,
						Status: v1.ConditionTrue,
					},
					{
						Type:   appsv1.DeploymentProgressing,
						Status: v1.ConditionTrue,
					},
				}
			})).Should(Succeed())

			Eventually(providerHelper.ToMatch(func(g Gomega, provider *pipelinesv1.Provider) {
				g.Expect(provider.Status.SynchronizationState).To(Equal(apis.Succeeded))
				// g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
			})).Should(Succeed())

			// Expect(pipelineHelper.Update(func(pipeline *pipelinesv1.Pipeline) {
			// 	pipeline.Spec = pipelinesv1.RandomPipelineSpec(Provider.Name)
			// })).To(Succeed())
			//
			// Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
			// 	g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Updating))
			// 	g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Updating))
			// })).Should(Succeed())
			//
			// Eventually(pipelineHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
			// 	workflow.Status.Phase = argo.WorkflowSucceeded
			// 	workflowutil.SetProviderOutput(workflow, providers.Output{Id: providerId})
			// })).Should(Succeed())
			//
			// Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
			// 	g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Succeeded))
			// 	g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
			// 	g.Expect(pipeline.Status.Provider.Name).To(Equal(pipeline.Spec.Provider))
			// })).Should(Succeed())
			//
			// Expect(pipelineHelper.Delete()).To(Succeed())
			//
			// Eventually(pipelineHelper.ToMatch(func(g Gomega, pipeline *pipelinesv1.Pipeline) {
			// 	g.Expect(pipeline.Status.SynchronizationState).To(Equal(apis.Deleting))
			// 	g.Expect(pipeline.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Deleting))
			// })).Should(Succeed())
			//
			// Eventually(pipelineHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
			// 	workflow.Status.Phase = argo.WorkflowSucceeded
			// 	workflowutil.SetProviderOutput(workflow, providers.Output{Id: ""})
			// })).Should(Succeed())
			//
			// Eventually(pipelineHelper.Exists).Should(Not(Succeed()))
			//
			// Eventually(pipelineHelper.EmittedEventsToMatch(func(g Gomega, events []v1.Event) {
			// 	g.Expect(events).To(ConsistOf(
			// 		HaveReason(EventReasons.Syncing),
			// 		HaveReason(EventReasons.Synced),
			// 		HaveReason(EventReasons.Syncing),
			// 		HaveReason(EventReasons.Synced),
			// 		HaveReason(EventReasons.Syncing),
			// 		HaveReason(EventReasons.Synced),
			// 	))
			// })).Should(Succeed())
		})
	})
})
