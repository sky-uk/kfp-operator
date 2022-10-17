//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"github.com/sky-uk/kfp-operator/providers/base"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Experiment controller k8s integration", Serial, func() {
	When("Creating, updating and deleting", func() {
		It("transitions through all stages", func() {
			experiment := pipelinesv1.RandomExperiment()

			kfpId := "12345"
			anotherKfpId := "67890"
			testCtx := NewExperimentTestContext(experiment, k8sClient, ctx)

			Expect(k8sClient.Create(ctx, testCtx.Experiment)).To(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(experiment.Status.ObservedGeneration).To(Equal(experiment.GetGeneration()))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(WorkflowConstants.CreateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, base.Output{Id: kfpId})
			})).Should(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(WorkflowConstants.CreateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.UpdateExperiment(func(pipeline *pipelinesv1.Experiment) {
				pipeline.Spec = pipelinesv1.RandomExperimentSpec()
			})).To(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Updating))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(WorkflowConstants.UpdateOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, base.Output{Id: anotherKfpId})
			})).Should(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Succeeded))
			})).Should(Succeed())
			Eventually(testCtx.FetchWorkflow(WorkflowConstants.UpdateOperationLabel)).Should(Not(Succeed()))

			Expect(testCtx.DeleteExperiment()).To(Succeed())

			Eventually(testCtx.ExperimentToMatch(func(g Gomega, experiment *pipelinesv1.Experiment) {
				g.Expect(experiment.Status.SynchronizationState).To(Equal(apis.Deleting))
			})).Should(Succeed())

			Eventually(testCtx.WorkflowToBeUpdated(WorkflowConstants.DeleteOperationLabel, func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, base.Output{Id: ""})
			})).Should(Succeed())

			Eventually(testCtx.ExperimentExists).Should(Not(Succeed()))
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
})
