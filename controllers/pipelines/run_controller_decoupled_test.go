//go:build decoupled
// +build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
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

	When("the completion state is set", func() {
		It("sets MarkCompletedAt", func() {
			runHelper := CreateSucceeded(pipelinesv1.RandomRun())

			Expect(runHelper.UpdateStatus(func(run *pipelinesv1.Run) {
				run.Status.CompletionState = pipelinesv1.CompletionStates.Succeeded
			})).To(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.MarkedCompletedAt).NotTo(BeNil())
			})).Should(Succeed())
		})
	})

	When("MarkCompletedAt is set and the TTL has passed", func() {
		It("deletes the resource", func() {
			runHelper := CreateSucceeded(pipelinesv1.RandomRun())

			Expect(runHelper.UpdateStatus(func(run *pipelinesv1.Run) {
				// time.Sub does not exist for Durations
				run.Status.MarkedCompletedAt = &metav1.Time{Time: time.Now().Add(-testConfig.RunCompletionTTL.Duration)}
			})).To(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Deleting))
			})).Should(Succeed())
		})
	})

	When("The pipeline version is fixed", func() {
		It("triggers a create with an ObservedPipelineVersion that matches the fixed version", func() {
			run := pipelinesv1.RandomRun()
			pipelineVersion := apis.RandomString()
			run.Spec.Pipeline = pipelinesv1.PipelineIdentifier{Name: apis.RandomString(), Version: pipelineVersion}

			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.ObservedPipelineVersion).To(Equal(pipelineVersion))
			})).Should(Succeed())
		})
	})

	When("The pipeline version is not fixed and the pipeline has succeeded", func() {
		It("triggers a create with an ObservedPipelineVersion that matches the current pipeline version", func() {
			pipeline := pipelinesv1.RandomPipeline()
			CreateSucceeded(pipeline)

			run := pipelinesv1.RandomRun()
			run.Spec.Pipeline = pipeline.UnversionedIdentifier()
			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.ObservedPipelineVersion).To(Equal(pipeline.Status.Version))
			})).Should(Succeed())
		})
	})

	When("The pipeline version is not fixed and the pipeline has not succeeded", func() {
		It("fails the run with an empty ObservedPipelineVersion", func() {
			pipeline := pipelinesv1.RandomPipeline()
			CreateStable(pipeline)

			run := pipelinesv1.RandomRun()
			run.Spec.Pipeline = pipeline.UnversionedIdentifier()
			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Failed))
				g.Expect(run.Status.ObservedPipelineVersion).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("The pipeline version is not fixed and the pipeline succeeds", func() {
		It("triggers a create with an ObservedPipelineVersion that matches the current pipeline version", func() {
			pipeline := pipelinesv1.RandomPipeline()
			pipelineHelper := CreateStable(pipeline)

			run := pipelinesv1.RandomRun()
			run.Spec.Pipeline = pipeline.UnversionedIdentifier()
			runHelper := Create(run)

			pipelineHelper.UpdateToSucceeded()

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.ObservedPipelineVersion).To(Equal(pipeline.Status.Version))
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration does not exist", func() {
		It("fails with the dependency unset", func() {
			runConfigurationName := apis.RandomString()
			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name: runConfigurationName,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			pipelineVersion := apis.RandomString()
			run.Spec.Pipeline = pipelinesv1.PipelineIdentifier{Name: apis.RandomString(), Version: pipelineVersion}

			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Failed))
				g.Expect(run.Status.Dependencies[runConfigurationName].ProviderId).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration has no succeeded run", func() {
		It("fails with the dependency unset", func() {
			referencedRc := pipelinesv1.RandomRunConfiguration()
			Expect(k8sClient.Create(ctx, referencedRc)).To(Succeed())
			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name: referencedRc.Name,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			pipelineVersion := apis.RandomString()
			run.Spec.Pipeline = pipelinesv1.PipelineIdentifier{Name: apis.RandomString(), Version: pipelineVersion}

			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Failed))
				g.Expect(run.Status.Dependencies[referencedRc.Name].ProviderId).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration has succeeded but misses the outputArtifact", func() {
		It("fails with the dependency missing the artifact", func() {
			artifactName := apis.RandomString()
			providerId := apis.RandomString()

			referencedRc := pipelinesv1.RandomRunConfiguration()
			Expect(k8sClient.Create(ctx, referencedRc)).To(Succeed())

			referencedRc.Status.LatestRuns.Succeeded.ProviderId = providerId
			Expect(k8sClient.Status().Update(ctx, referencedRc)).To(Succeed())

			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					Name: artifactName,
					ValueFrom: pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name: referencedRc.Name,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			pipelineVersion := apis.RandomString()
			run.Spec.Pipeline = pipelinesv1.PipelineIdentifier{Name: apis.RandomString(), Version: pipelineVersion}

			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Failed))
				g.Expect(run.Status.Dependencies[referencedRc.Name].ProviderId).To(Equal(providerId))
				g.Expect(run.Status.Dependencies[referencedRc.Name].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration has succeeded with the outputArtifact", func() {
		It("triggers a create with the artifact set", func() {
			artifactName := apis.RandomString()
			providerId := apis.RandomString()
			artifact := common.Artifact{
				Name: artifactName,
				Location: apis.RandomString(),
			}

			referencedRc := pipelinesv1.RandomRunConfiguration()
			Expect(k8sClient.Create(ctx, referencedRc)).To(Succeed())

			referencedRc.Status.LatestRuns.Succeeded.Artifacts = []common.Artifact{artifact}
			referencedRc.Status.LatestRuns.Succeeded.ProviderId = providerId
			Expect(k8sClient.Status().Update(ctx, referencedRc)).To(Succeed())

			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					Name: artifactName,
					ValueFrom: pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name: referencedRc.Name,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			pipelineVersion := apis.RandomString()
			run.Spec.Pipeline = pipelinesv1.PipelineIdentifier{Name: apis.RandomString(), Version: pipelineVersion}

			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.Dependencies[referencedRc.Name].ProviderId).To(Equal(providerId))
				g.Expect(run.Status.Dependencies[referencedRc.Name].Artifacts).To(ContainElement(artifact))
			})).Should(Succeed())
		})
	})
})
