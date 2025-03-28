//go:build decoupled

package pipelines

import (
	"time"

	argo "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	providers "github.com/sky-uk/kfp-operator/argo/providers/base"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Run controller k8s integration", Serial, func() {
	When("Creating and deleting", func() {
		It("transitions through all stages", func() {
			providerId := "12345"
			runHelper := Create(pipelineshub.RandomRun(Provider.GetCommonNamespacedName()))

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelineshub.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Creating))
				g.Expect(run.Status.ObservedGeneration).To(Equal(run.GetGeneration()))
			})).Should(Succeed())

			Eventually(runHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				workflowutil.SetProviderOutput(workflow, providers.Output{Id: providerId})
			})).Should(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelineshub.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Succeeded))
				g.Expect(run.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
				g.Expect(run.Status.Provider.Name).To(Equal(run.Spec.Provider))
			})).Should(Succeed())

			Expect(runHelper.Update(func(run *pipelineshub.Run) {
				run.Spec = pipelineshub.RandomRunSpec(Provider.GetCommonNamespacedName())
			})).To(MatchError(ContainSubstring("immutable")))

			Expect(runHelper.Delete()).To(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelineshub.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Deleting))
				g.Expect(run.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Deleting))
			})).Should(Succeed())

			Eventually(runHelper.WorkflowToBeUpdated(func(workflow *argo.Workflow) {
				workflow.Status.Phase = argo.WorkflowSucceeded
				workflowutil.SetProviderOutput(workflow, providers.Output{Id: ""})
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
			run := pipelineshub.RandomRun(Provider.GetCommonNamespacedName())
			run.Spec.RuntimeParameters = []pipelineshub.RuntimeParameter{
				{
					Value: apis.RandomString(),
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           apis.RandomString(),
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			Expect(K8sClient.Create(Ctx, run)).To(MatchError(ContainSubstring("only one of value or valueFrom can be set")))
		})
	})

	When("the completion state is set", func() {
		It("sets MarkCompletedAt", func() {
			runHelper := CreateSucceeded(pipelineshub.RandomRun(Provider.GetCommonNamespacedName()))

			Expect(runHelper.UpdateStatus(func(run *pipelineshub.Run) {
				run.Status.CompletionState = pipelineshub.CompletionStates.Succeeded
			})).To(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelineshub.Run) {
				g.Expect(run.Status.MarkedCompletedAt).NotTo(BeNil())
			})).Should(Succeed())
		})
	})

	When("MarkCompletedAt is set and the TTL has passed", func() {
		It("deletes the resource", func() {
			runHelper := CreateSucceeded(pipelineshub.RandomRun(Provider.GetCommonNamespacedName()))

			Expect(runHelper.UpdateStatus(func(run *pipelineshub.Run) {
				// time.Sub does not exist for Durations
				run.Status.MarkedCompletedAt = &metav1.Time{Time: time.Now().Add(-TestConfig.RunCompletionTTL.Duration)}
			})).To(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelineshub.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Deleting))
			})).Should(Succeed())
		})
	})

	When("The pipeline version is fixed", func() {
		It("triggers a create with an ObservedPipelineVersion that matches the fixed version", func() {
			run := pipelineshub.RandomRun(Provider.GetCommonNamespacedName())
			pipelineVersion := apis.RandomString()
			run.Spec.Pipeline = pipelineshub.PipelineIdentifier{
				Name:    apis.RandomString(),
				Version: pipelineVersion,
			}

			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelineshub.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.ObservedPipelineVersion).To(Equal(pipelineVersion))
			})).Should(Succeed())
		})
	})

	When("The pipeline version is not fixed and the pipeline has succeeded", func() {
		It("triggers a create with an ObservedPipelineVersion that matches the current pipeline version", func() {
			pipeline := pipelineshub.RandomPipeline(Provider.GetCommonNamespacedName())
			pipeline.Spec.Framework.Type = TestFramework
			CreateSucceeded(pipeline)

			run := pipelineshub.RandomRun(Provider.GetCommonNamespacedName())
			run.Spec.Pipeline = pipeline.UnversionedIdentifier()
			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelineshub.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.ObservedPipelineVersion).To(Equal(pipeline.Status.Version))
			})).Should(Succeed())
		})
	})

	When("The pipeline version is not fixed and the pipeline succeeds", func() {
		It("triggers a create with an ObservedPipelineVersion that matches the current pipeline version", func() {
			pipeline := pipelineshub.RandomPipeline(Provider.GetCommonNamespacedName())
			pipeline.Spec.Framework.Type = TestFramework
			pipelineHelper := CreateStable(pipeline)

			run := pipelineshub.RandomRun(Provider.GetCommonNamespacedName())
			run.Spec.Pipeline = pipeline.UnversionedIdentifier()
			runHelper := Create(run)

			pipelineHelper.UpdateToSucceeded()

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelineshub.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.ObservedPipelineVersion).To(Equal(pipeline.Status.Version))
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration does not exist", func() {
		It("unsets the dependency", func() {
			runConfigurationName := apis.RandomString()
			run := pipelineshub.RandomRun(Provider.GetCommonNamespacedName())
			run.Spec.RuntimeParameters = []pipelineshub.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           runConfigurationName,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			runHelper := Create(run)

			oldState := run.Status.SynchronizationState
			Eventually(runHelper.ToMatch(func(g Gomega, fetchedRun *pipelineshub.Run) {
				g.Expect(fetchedRun.Status.SynchronizationState).To(Equal(oldState))
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[runConfigurationName].ProviderId).To(BeEmpty())
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[runConfigurationName].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration has no succeeded run", func() {
		It("unsets the dependency", func() {
			referencedRc := pipelineshub.RandomRunConfiguration(Provider.GetCommonNamespacedName())
			Expect(K8sClient.Create(Ctx, referencedRc)).To(Succeed())

			run := pipelineshub.RandomRun(Provider.GetCommonNamespacedName())
			run.Spec.RuntimeParameters = []pipelineshub.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           referencedRc.Name,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			runHelper := Create(run)

			oldState := run.Status.SynchronizationState
			Eventually(runHelper.ToMatch(func(g Gomega, fetchedRun *pipelineshub.Run) {
				g.Expect(fetchedRun.Status.SynchronizationState).To(Equal(oldState))
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[referencedRc.Name].ProviderId).To(BeEmpty())
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[referencedRc.Name].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration has succeeded but misses the outputArtifact", func() {
		It("unsets the dependency", func() {
			referencedRc := createRcWithLatestRun(pipelineshub.RunReference{
				ProviderId: apis.RandomString(),
				Artifacts:  []common.Artifact{common.RandomArtifact()},
			})

			run := pipelineshub.RandomRun(Provider.GetCommonNamespacedName())
			run.Spec.RuntimeParameters = []pipelineshub.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           referencedRc.Name,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			runHelper := Create(run)

			oldState := run.Status.SynchronizationState
			Eventually(runHelper.ToMatch(func(g Gomega, fetchedRun *pipelineshub.Run) {
				g.Expect(fetchedRun.Status.SynchronizationState).To(Equal(oldState))
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[referencedRc.Name].ProviderId).To(BeEmpty())
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations[referencedRc.Name].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A RunConfiguration reference has been removed", func() {
		It("removes the dependency", func() {
			run := pipelineshub.RandomRun(Provider.GetCommonNamespacedName())
			run.Spec.RuntimeParameters = []pipelineshub.RuntimeParameter{}
			runHelper := Create(run)

			excessDependency := apis.RandomString()
			run.SetDependencyRuns(map[string]pipelineshub.RunReference{excessDependency: {}})
			Expect(K8sClient.Status().Update(Ctx, run)).To(Succeed())

			Eventually(runHelper.ToMatch(func(g Gomega, fetchedRun *pipelineshub.Run) {
				g.Expect(fetchedRun.Status.Dependencies.RunConfigurations).NotTo(HaveKey(excessDependency))
			})).Should(Succeed())
		})
	})

	When("Referenced RunConfigurations has succeeded with the outputArtifacts", func() {
		It("triggers a create with the artifacts set", func() {
			referencedRc1 := createRcWithLatestRun(pipelineshub.RunReference{
				ProviderId: apis.RandomString(),
				Artifacts: []common.Artifact{
					common.RandomArtifact(),
					common.RandomArtifact(),
				},
			})

			referencedRc2 := createRcWithLatestRun(pipelineshub.RunReference{
				ProviderId: apis.RandomString(),
				Artifacts: []common.Artifact{
					common.RandomArtifact(),
				},
			})

			run := pipelineshub.RandomRun(Provider.GetCommonNamespacedName())
			run.Spec.RuntimeParameters = []pipelineshub.RuntimeParameter{
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           referencedRc1.Name,
							OutputArtifact: referencedRc1.Status.LatestRuns.Succeeded.Artifacts[0].Name,
						},
					},
				},
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           referencedRc1.Name,
							OutputArtifact: referencedRc1.Status.LatestRuns.Succeeded.Artifacts[1].Name,
						},
					},
				},
				{
					Name: apis.RandomString(),
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           referencedRc2.Name,
							OutputArtifact: referencedRc2.Status.LatestRuns.Succeeded.Artifacts[0].Name,
						},
					},
				},
			}

			runHelper := Create(run)

			Eventually(runHelper.ToMatch(func(g Gomega, run *pipelineshub.Run) {
				g.Expect(run.Status.SynchronizationState).To(Equal(apis.Creating))
				g.Expect(run.Status.Dependencies.RunConfigurations[referencedRc1.Name]).To(Equal(referencedRc1.Status.LatestRuns.Succeeded))
				g.Expect(run.Status.Dependencies.RunConfigurations[referencedRc2.Name]).To(Equal(referencedRc2.Status.LatestRuns.Succeeded))
			})).Should(Succeed())
		})
	})
})
