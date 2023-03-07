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

var _ = Describe("Run controller k8s integration", Serial, func() {
	When("Creating and deleting", func() {
		It("transitions through all stages", func() {
			providerId := "12345"
			runHelper := Create(pipelinesv1.RandomRun())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.ObservedGeneration).To(Equal(run.GetGeneration()))
			})).Should(Succeed())

			Eventually(runHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(run.Status.ProviderId.Provider).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())

			Expect(runHelper.Update(func(run *pipelinesv1.Run) {
				run.Spec = pipelinesv1.RandomRunSpec()
			})).To(MatchError(ContainSubstring("immutable")))

			Expect(runHelper.Delete()).To(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Deleting))
			})).Should(Succeed())

			Eventually(runHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: ""})
			})).Should(Succeed())

			Eventually(runHelper.Exists).Should(Not(Succeed()))

			Eventually(runHelper.EmittedEventsToMatch(func(g Gomega, events []v1.Event) {
				g.Expect(events).To(ConsistOf(
					HaveReason(EventReasons.Syncing),
					HaveReason(EventReasons.Synced),
					HaveReason(EventReasons.Syncing),
					HaveReason(EventReasons.Synced),
				))
			})).Should(Succeed())
		})
	})
})
