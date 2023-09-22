//go:build decoupled

package pipelines

import (
	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
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
				g.Expect(run.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Creating))
				g.Expect(run.Status.ObservedGeneration).To(Equal(run.GetGeneration()))
			})).Should(Succeed())

			Eventually(runHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				setProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(run.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(run.Status.ProviderId.Provider).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())

			Expect(runHelper.Update(func(run *pipelinesv1.Run) {
				run.Spec = pipelinesv1.RandomRunSpec()
			})).To(MatchError(ContainSubstring("immutable")))

			Expect(runHelper.Delete()).To(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Deleting))
				g.Expect(run.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Deleting))
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

	When("Creating an invalid run", func() {
		It("errors", func() {
			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					Value: apis.RandomString(),
					ValueFrom: &pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name:           apis.RandomString(),
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			Expect(k8sClient.Create(ctx, run)).To(MatchError(ContainSubstring("only one of value or valueFrom can be set")))
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
				g.Expect(run.Status.Dependencies.Pipeline.Version).To(Equal(pipelineVersion))
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
				g.Expect(run.Status.Dependencies.Pipeline.Version).To(Equal(pipeline.Status.Version))
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
				g.Expect(run.Status.Dependencies.Pipeline.Version).To(Equal(pipeline.Status.Version))
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration does not exist", func() {
		It("unsets the dependency", func() {
			runConfigurationName := apis.RandomString()
			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name:           runConfigurationName,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			runHelper := Create(run)

			oldState := run.Status.SynchronizationState
			Eventually(runHelper.ToMatch(func(g Gomega, fetchedRun *pipelinesv1.Run) {
				g.Expect(fetchedRun.Status.SynchronizationState).To(Equal(oldState))
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[runConfigurationName].ProviderId).To(BeEmpty())
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[runConfigurationName].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration has no succeeded run", func() {
		It("unsets the dependency", func() {
			referencedRc := pipelinesv1.RandomRunConfiguration()
			Expect(k8sClient.Create(ctx, referencedRc)).To(Succeed())

			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name:           referencedRc.Name,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			runHelper := Create(run)

			oldState := run.Status.SynchronizationState
			Eventually(runHelper.ToMatch(func(g Gomega, fetchedRun *pipelinesv1.Run) {
				g.Expect(fetchedRun.Status.SynchronizationState).To(Equal(oldState))
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[referencedRc.Name].ProviderId).To(BeEmpty())
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[referencedRc.Name].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration has succeeded but misses the outputArtifact", func() {
		It("unsets the dependency", func() {
			referencedRc := createRcWithLatestRun(pipelinesv1.RunReference{
				ProviderId: apis.RandomString(),
				Artifacts:  []common.Artifact{common.RandomArtifact()},
			})

			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name:           referencedRc.Name,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			runHelper := Create(run)

			oldState := run.Status.SynchronizationState
			Eventually(runHelper.ToMatch(func(g Gomega, fetchedRun *pipelinesv1.Run) {
				g.Expect(fetchedRun.Status.SynchronizationState).To(Equal(oldState))
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[referencedRc.Name].ProviderId).To(BeEmpty())
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[referencedRc.Name].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A RunConfiguration reference has been removed", func() {
		It("removes the dependency", func() {
			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{}
			runHelper := Create(run)

			excessDependency := apis.RandomString()
			run.SetDependencyRuns(map[string]pipelinesv1.RunReference{excessDependency: {}})
			Expect(k8sClient.Status().Update(ctx, run)).To(Succeed())

			oldState := run.Status.SynchronizationState
			Eventually(runHelper.ToMatch(func(g Gomega, fetchedRun *pipelinesv1.Run) {
				g.Expect(fetchedRun.Status.SynchronizationState).To(Equal(oldState))
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations).NotTo(HaveKey(excessDependency))
			})).Should(Succeed())
		})
	})

	When("Referenced RunConfigurations has succeeded with the outputArtifacts", func() {
		It("triggers a create with the artifacts set", func() {
			referencedRc1 := createRcWithLatestRun(pipelinesv1.RunReference{
				ProviderId: apis.RandomString(),
				Artifacts: []common.Artifact{
					common.RandomArtifact(),
					common.RandomArtifact(),
				},
			})

			referencedRc2 := createRcWithLatestRun(pipelinesv1.RunReference{
				ProviderId: apis.RandomString(),
				Artifacts: []common.Artifact{
					common.RandomArtifact(),
				},
			})

			run := pipelinesv1.RandomRun()
			run.Spec.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name:           referencedRc1.Name,
							OutputArtifact: referencedRc1.Status.LatestRuns.Succeeded.Artifacts[0].Name,
						},
					},
				},
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name:           referencedRc1.Name,
							OutputArtifact: referencedRc1.Status.LatestRuns.Succeeded.Artifacts[1].Name,
						},
					},
				},
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name:           referencedRc2.Name,
							OutputArtifact: referencedRc2.Status.LatestRuns.Succeeded.Artifacts[0].Name,
						},
					},
				},
			}

			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelinesv1.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.Dependencies.RunConfigurations[referencedRc1.Name]).To(Equal(referencedRc1.Status.LatestRuns.Succeeded))
				g.Expect(run.Status.Dependencies.RunConfigurations[referencedRc2.Name]).To(Equal(referencedRc2.Status.LatestRuns.Succeeded))
			})).Should(Succeed())
		})
	})
})

func createRcWithLatestRun(succeeded pipelinesv1.RunReference) *pipelinesv1.RunConfiguration {
	referencedRc := pipelinesv1.RandomRunConfiguration()
	referencedRc.Spec.Triggers = pipelinesv1.Triggers{}
	Expect(k8sClient.Create(ctx, referencedRc)).To(Succeed())
	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, referencedRc.GetNamespacedName(), referencedRc)).To(Succeed())
		g.Expect(referencedRc.Status.SynchronizationState).To(Equal(apis.Succeeded))
	}).Should(Succeed())
	referencedRc.Status.LatestRuns.Succeeded = succeeded
	Expect(k8sClient.Status().Update(ctx, referencedRc)).To(Succeed())

	return referencedRc
}
