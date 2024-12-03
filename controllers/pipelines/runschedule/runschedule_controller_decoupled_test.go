//go:build decoupled

package runschedule

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	"github.com/sky-uk/kfp-operator/controllers/pipelines"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("RunSchedule controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			providerId := apis.RandomString()
			rcHelper := pipelines.Create(pipelinesv1.RandomRunSchedule(testutil.Provider.Name))
			Eventually(rcHelper.ToMatch(func(g Gomega, runSchedule *pipelinesv1.RunSchedule) {
				g.Expect(runSchedule.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(runSchedule.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Creating))
				g.Expect(runSchedule.Status.ObservedGeneration).To(Equal(runSchedule.GetGeneration()))
			})).Should(Succeed())

			Eventually(rcHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				pipelines.SetProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(rcHelper.ToMatch(func(g Gomega, runSchedule *pipelinesv1.RunSchedule) {
				g.Expect(runSchedule.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(runSchedule.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(runSchedule.Status.Provider.Name).To(Equal(runSchedule.Spec.Provider))
			})).Should(Succeed())

			Expect(rcHelper.Update(func(runSchedule *pipelinesv1.RunSchedule) {
				runSchedule.Spec = pipelinesv1.RandomRunScheduleSpec(testutil.Provider.Name)
			})).To(Succeed())

			Eventually(rcHelper.ToMatch(func(g Gomega, runSchedule *pipelinesv1.RunSchedule) {
				g.Expect(runSchedule.Status.SynchronizationState).To(Equal(apis.Updating))
				g.Expect(runSchedule.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Updating))
			})).Should(Succeed())

			Eventually(rcHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				pipelines.SetProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(rcHelper.ToMatch(func(g Gomega, runSchedule *pipelinesv1.RunSchedule) {
				g.Expect(runSchedule.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(runSchedule.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(runSchedule.Status.Provider.Name).To(Equal(runSchedule.Spec.Provider))
			})).Should(Succeed())

			Expect(rcHelper.Delete()).To(Succeed())

			Eventually(rcHelper.ToMatch(func(g Gomega, runSchedule *pipelinesv1.RunSchedule) {
				g.Expect(runSchedule.Status.SynchronizationState).To(Equal(apis.Deleting))
				g.Expect(runSchedule.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Deleting))
			})).Should(Succeed())

			Eventually(rcHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				pipelines.SetProviderOutput(workflow, providers.Output{Id: ""})
			})).Should(Succeed())

			Eventually(rcHelper.Exists).Should(Not(Succeed()))

			Eventually(rcHelper.EmittedEventsToMatch(func(g Gomega, events []v1.Event) {
				g.Expect(events).To(ConsistOf(
					testutil.HaveReason(pipelines.EventReasons.Syncing),
					testutil.HaveReason(pipelines.EventReasons.Synced),
					testutil.HaveReason(pipelines.EventReasons.Syncing),
					testutil.HaveReason(pipelines.EventReasons.Synced),
					testutil.HaveReason(pipelines.EventReasons.Syncing),
					testutil.HaveReason(pipelines.EventReasons.Synced),
				))
			})).Should(Succeed())
		})
	})
})
