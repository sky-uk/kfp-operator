//go:build decoupled

package pipelines

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"github.com/sky-uk/kfp-operator/common"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
)

var _ = Describe("RunConfiguration controller k8s integration", Serial, func() {
	It("creates RunSchedule owned resources that are in the triggers", func() {
		runConfiguration := createStableRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
			runConfiguration.Spec.Triggers = pipelinesv1.RandomScheduleTrigger()
			return runConfiguration
		}, apis.Updating)

		Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelinesv1.RunSchedule) {
			//other fields tested in unit test
			g.Expect(ownedSchedule.Spec.Schedule).To(Equal(runConfiguration.Spec.Triggers.Schedules[0]))
			g.Expect(ownedSchedule.Status.SynchronizationState).To(Equal(apis.Creating))
		})).Should(Succeed())

		Expect(updateOwnedSchedules(runConfiguration, func(ownedSchedule *pipelinesv1.RunSchedule) {
			ownedSchedule.Status.SynchronizationState = apis.Succeeded
		})).To(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(K8sClient.Get(Ctx, runConfiguration.GetNamespacedName(), runConfiguration)).To(Succeed())
			g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Succeeded))
			g.Expect(runConfiguration.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
			g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
		}).Should(Succeed())
	})

	It("deletes RunSchedule owned resources that are not in the triggers", func() {
		runConfiguration := createSucceededRcWithSchedule()

		runConfiguration.Spec.Triggers.Schedules = nil
		Expect(K8sClient.Update(Ctx, runConfiguration)).To(Succeed())

		Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
			g.Expect(fetchedRc.Status.SynchronizationState).To(Equal(apis.Succeeded))
			g.Expect(fetchedRc.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
			g.Expect(fetchedRc.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
		})).Should(Succeed())

		Eventually(hasNoSchedules(runConfiguration)).Should(Succeed())
	})

	It("succeeds without creating resources if dependencies are not met", func() {

		runConfiguration := createSucceededRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
			runConfiguration.Spec.Triggers = pipelinesv1.RandomScheduleTrigger()
			runConfiguration.Spec.Triggers.OnChange = []pipelinesv1.OnChangeType{pipelinesv1.OnChangeTypes.RunSpec}
			runConfiguration.Spec.Run.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				pipelinesv1.RandomRunConfigurationRefRuntimeParameter(),
			}
			return runConfiguration
		})

		Eventually(hasNoSchedules(runConfiguration)).Should(Succeed())
		Eventually(hasNoRuns(runConfiguration)).Should(Succeed())
	})

	When("Deleted", func() {
		It("cascades deletes", func() {
			Skip("See https://github.com/kubernetes-sigs/controller-runtime/issues/1459. Keep test for documentation")
			runConfiguration := createSucceededRcWithSchedule()

			Expect(K8sClient.Delete(Ctx, runConfiguration)).To(Succeed())

			Eventually(hasNoSchedules(runConfiguration)).Should(Succeed())
		})
	})

	When("Creating an RC with a fixed pipeline version", func() {
		It("sets the ObservedPipelineVersion to the fixed version", func() {
			pipelineVersion := apis.RandomString()

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = pipelinesv1.PipelineIdentifier{Name: apis.RandomString(), Version: pipelineVersion}
				return runConfiguration
			})

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.ObservedPipelineVersion).To(Equal(pipelineVersion))
			})).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with no version specified on the RC", func() {
		It("triggers an update of the run configuration", func() {
			pipeline := pipelinesv1.RandomPipeline(Provider.Name)
			pipelineHelper := CreateSucceeded(pipeline)

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
				runConfiguration.Status.ObservedPipelineVersion = pipeline.ComputeVersion()
				return runConfiguration
			})

			pipelineHelper.UpdateStable(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec(Provider.Name)
			})

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.ObservedPipelineVersion).To(Equal(pipeline.ComputeVersion()))
			})).Should(Succeed())
		})

		It("change trigger creates a run when the pipeline is updated", func() {
			pipeline := pipelinesv1.RandomPipeline(Provider.Name)
			pipelineHelper := CreateSucceeded(pipeline)
			firstPipelineVersion := pipeline.ComputeVersion()

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
				runConfiguration.Spec.Triggers = pipelinesv1.RandomOnChangeTrigger()
				return runConfiguration
			})

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(ConsistOf(HavePipelineVersion(firstPipelineVersion)))
			}).Should(Succeed())

			pipelineHelper.UpdateStable(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec(Provider.Name)
			})

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(ConsistOf(HavePipelineVersion(firstPipelineVersion), HavePipelineVersion(pipeline.ComputeVersion())))
			}).Should(Succeed())
		})

		It("change trigger creates a run when the pipeline is created", func() {
			pipeline := pipelinesv1.RandomPipeline(Provider.Name)

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
				runConfiguration.Spec.Triggers = pipelinesv1.RandomOnChangeTrigger()
				return runConfiguration
			})

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
			}))
			ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
			Expect(err).NotTo(HaveOccurred())
			Expect(ownedRuns).To(BeEmpty())

			CreateSucceeded(pipeline)

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(ConsistOf(HavePipelineVersion(pipeline.ComputeVersion())))
			}).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with a fixed version specified on the RC", func() {
		It("does not trigger an update of the run configuration", func() {
			pipeline := pipelinesv1.RandomPipeline(Provider.Name)
			fixedIdentifier := pipeline.VersionedIdentifier()

			pipelineHelper := CreateSucceeded(pipeline)

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = fixedIdentifier
				runConfiguration.Status.ObservedPipelineVersion = pipeline.ComputeVersion()
				return runConfiguration
			})

			pipelineHelper.UpdateStable(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec(Provider.Name)
			})

			// To verify the absence of additional RC updates, force another update of the resource.
			// If the update is processed but the pipeline version hasn't changed,
			// given that reconciliation requests are processed in-order, we can conclude that the RC is fixed.
			newExperiment := apis.RandomString()

			runConfiguration.Spec.Run.ExperimentName = newExperiment
			Expect(K8sClient.Update(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Spec.Run.ExperimentName).To(Equal(newExperiment))
				g.Expect(fetchedRc.Status.ObservedPipelineVersion).To(Equal(fixedIdentifier.Version))
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration does not exist", func() {
		It("unsets the dependency", func() {
			runConfigurationName := apis.RandomString()

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
				runConfiguration.Spec.Run.RuntimeParameters = []pipelinesv1.RuntimeParameter{
					{
						ValueFrom: &pipelinesv1.ValueFrom{
							RunConfigurationRef: pipelinesv1.RunConfigurationRef{
								Name:           runConfigurationName,
								OutputArtifact: apis.RandomString(),
							},
						},
					},
				}
				return runConfiguration
			})

			runConfiguration.Status.Dependencies.RunConfigurations = map[string]pipelinesv1.RunReference{
				runConfigurationName: {
					ProviderId: apis.RandomString(),
					Artifacts:  []common.Artifact{common.RandomArtifact()},
				},
			}

			Expect(K8sClient.Status().Update(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.ObservedGeneration).To(Equal(fetchedRc.Generation))
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[runConfigurationName].ProviderId).To(BeEmpty())
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[runConfigurationName].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration has succeeded but misses the outputArtifact", func() {
		It("unsets the dependency", func() {
			referencedRc := createRcWithLatestRun(pipelinesv1.RunReference{
				ProviderId: apis.RandomString(),
			})

			runConfiguration := pipelinesv1.RandomRunConfiguration(Provider.Name)
			runConfiguration.Spec.Run.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					ValueFrom: &pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name:           referencedRc.Name,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			oldState := runConfiguration.Status.SynchronizationState
			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.SynchronizationState).To(Equal(oldState))
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[referencedRc.Name].ProviderId).To(BeEmpty())
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[referencedRc.Name].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A RunConfiguration reference has been removed", func() {
		It("removes the dependency", func() {
			runConfiguration := createSucceededRc()

			excessDependency := apis.RandomString()
			runConfiguration.SetDependencyRuns(map[string]pipelinesv1.RunReference{excessDependency: {}})

			Expect(K8sClient.Status().Update(Ctx, runConfiguration)).To(Succeed())

			oldState := runConfiguration.Status.SynchronizationState
			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.SynchronizationState).To(Equal(oldState))
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations).NotTo(HaveKey(excessDependency))
			})).Should(Succeed())
		})
	})

	When("Completing the referenced run configuration", func() {
		It("Sets the run configuration's dependency field", func() {
			referencedRc := createRcWithLatestRun(pipelinesv1.RunReference{
				ProviderId: apis.RandomString(),
				Artifacts:  []common.Artifact{common.RandomArtifact()},
			})

			runConfiguration := pipelinesv1.RandomRunConfiguration(Provider.Name)
			runConfiguration.Spec.Run.RuntimeParameters = []pipelinesv1.RuntimeParameter{
				{
					ValueFrom: &pipelinesv1.ValueFrom{
						RunConfigurationRef: pipelinesv1.RunConfigurationRef{
							Name:           referencedRc.Name,
							OutputArtifact: referencedRc.Status.LatestRuns.Succeeded.Artifacts[0].Name,
						},
					},
				},
			}

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			Expect(K8sClient.Get(Ctx, referencedRc.GetNamespacedName(), referencedRc)).To(Succeed())
			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[referencedRc.Name]).To(Equal(referencedRc.Status.LatestRuns.Succeeded))
			})).Should(Succeed())
		})

		It("RC trigger creates a run when the referenced RC has finished", func() {
			referencedRc := createRcWithLatestRun(pipelinesv1.RunReference{
				ProviderId: apis.RandomString(),
			})

			runConfiguration := pipelinesv1.RandomRunConfiguration(Provider.Name)
			runConfiguration.Spec.Triggers = pipelinesv1.Triggers{
				RunConfigurations: []string{
					referencedRc.Name,
				},
			}

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.Dependencies.RunConfigurations[referencedRc.Name].ProviderId).To(Equal(referencedRc.Status.LatestRuns.Succeeded.ProviderId))
			})).Should(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.Triggers.RunConfigurations[referencedRc.Name].ProviderId).To(Equal(referencedRc.Status.LatestRuns.Succeeded.ProviderId))
			})).Should(Succeed())

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(HaveLen(1))
			}).Should(Succeed())
		})
	})

	When("Removing a dependency trigger", func() {
		It("Removes a previously set triggers field", func() {
			referencedRc := createRcWithLatestRun(pipelinesv1.RunReference{
				ProviderId: apis.RandomString(),
			})

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
				runConfiguration.Spec.Triggers = pipelinesv1.Triggers{
					RunConfigurations: []string{
						referencedRc.Name,
					},
				}
				return runConfiguration
			})

			runConfiguration.Spec.Triggers.RunConfigurations = nil
			Expect(K8sClient.Update(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.Triggers.RunConfigurations).NotTo(HaveKey(referencedRc.Name))
			})).Should(Succeed())
		})
	})

	When("RunSpec changes", func() {
		It("onChange trigger creates a run when run template has changed", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration(Provider.Name)
			runConfiguration.Spec.Triggers = pipelinesv1.Triggers{
				OnChange: []pipelinesv1.OnChangeType{
					pipelinesv1.OnChangeTypes.RunSpec,
				},
			}

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Succeeded))
			})).Should(Succeed())

			runConfiguration.Spec.Run = pipelinesv1.RandomRunSpec(Provider.Name)

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(HaveLen(1))
			}).Should(Succeed())
		})
	})

	When("setting the provider", func() {
		It("stores the provider in the status", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration(Provider.Name)
			runConfiguration.Spec.Run.Provider = Provider.Name
			runConfiguration.Spec.Triggers = pipelinesv1.Triggers{}
			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.Provider).To(Equal(runConfiguration.Spec.Run.Provider))
			})).Should(Succeed())
		})

		It("passes the provider to owned resources", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration(Provider.Name)
			runConfiguration.Spec.Run.Provider = Provider.Name
			runConfiguration.Spec.Triggers = pipelinesv1.Triggers{
				Schedules: []pipelinesv1.Schedule{pipelinesv1.RandomSchedule()},
				OnChange:  []pipelinesv1.OnChangeType{pipelinesv1.OnChangeTypes.Pipeline},
			}
			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelinesv1.RunSchedule) {
				g.Expect(ownedSchedule.Spec.Provider).To(Equal(runConfiguration.Spec.Run.Provider))
			})).Should(Succeed())
			Eventually(matchRuns(runConfiguration, func(g Gomega, ownedRun *pipelinesv1.Run) {
				g.Expect(ownedRun.Spec.Provider).To(Equal(runConfiguration.Spec.Run.Provider))
			})).Should(Succeed())
		})
	})

	When("changing the provider", func() {
		It("fails the resource", func() {
			runConfiguration := createSucceededRcWithSchedule()
			runConfiguration.Spec.Run.Provider = apis.RandomLowercaseString()
			Expect(K8sClient.Update(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
				g.Expect(fetchedRc.Status.SynchronizationState).To(Equal(apis.Failed))
			})).Should(Succeed())

			Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelinesv1.RunSchedule) {
				g.Expect(ownedSchedule.Spec.Provider).To(Equal(Provider.Name))
			})).Should(Succeed())
		})
	})

	When("Creating an invalid run configuration", func() {
		It("errors", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration(Provider.Name)
			runConfiguration.Spec.Run.RuntimeParameters = []pipelinesv1.RuntimeParameter{
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

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(MatchError(ContainSubstring("only one of value or valueFrom can be set")))
		})
	})

	When("Updating an invalid run configuration", func() {
		It("errors", func() {
			runConfiguration := createSucceededRc()
			runConfiguration.Spec.Run.RuntimeParameters = []pipelinesv1.RuntimeParameter{
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

			Expect(K8sClient.Update(Ctx, runConfiguration)).To(MatchError(ContainSubstring("only one of value or valueFrom can be set")))
		})
	})
})

func matchRuns(runConfiguration *pipelinesv1.RunConfiguration, matcher func(Gomega, *pipelinesv1.Run)) func(Gomega) {
	return func(g Gomega) {
		ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedRuns).NotTo(BeEmpty())
		for _, ownedRun := range ownedRuns {
			matcher(g, &ownedRun)
		}
	}
}

func hasNoRuns(runConfiguration *pipelinesv1.RunConfiguration) func(Gomega) {
	return func(g Gomega) {
		ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedRuns).To(BeEmpty())
	}
}

func hasNoSchedules(runConfiguration *pipelinesv1.RunConfiguration) func(Gomega) {
	return func(g Gomega) {
		ownedSchedules, err := findOwnedRunSchedules(Ctx, K8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedSchedules).To(BeEmpty())
	}
}

func HavePipelineVersion(version string) types.GomegaMatcher {
	return &HavePipelineVersionMatcher{
		Version: version,
	}
}

type HavePipelineVersionMatcher struct {
	Version string
}

func (matcher *HavePipelineVersionMatcher) Match(actual interface{}) (success bool, err error) {
	run, isRun := actual.(pipelinesv1.Run)
	if !isRun {
		return false, fmt.Errorf("HavePipelineVersionMatcher matcher expects a Run.  Got:\n%s", format.Object(actual, 1))
	}

	return run.Spec.Pipeline.Version == matcher.Version, nil
}

func (matcher *HavePipelineVersionMatcher) FailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "to have version", matcher.Version)
}

func (matcher *HavePipelineVersionMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return format.Message(actual, "not to have version", matcher.Version)
}
