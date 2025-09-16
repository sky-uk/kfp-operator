//go:build decoupled

package pipelines

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
	"github.com/sky-uk/kfp-operator/pkg/common"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("RunConfiguration controller k8s integration", Serial, func() {
	It("creates RunSchedule owned resources that are in the triggers", func() {
		runConfiguration := createStableRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
			runConfiguration.Spec.Triggers = pipelineshub.RandomScheduleTrigger()
			return runConfiguration
		}, apis.Updating)

		Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelineshub.RunSchedule) {
			//other fields tested in unit test
			g.Expect(ownedSchedule.Spec.Schedule).To(Equal(runConfiguration.Spec.Triggers.Schedules[0]))
			g.Expect(ownedSchedule.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Creating))
		})).Should(Succeed())

		Expect(updateOwnedSchedules(runConfiguration, func(ownedSchedule *pipelineshub.RunSchedule) {
			ownedSchedule.Status.Conditions = ownedSchedule.Status.Conditions.SetReasonForSyncState(apis.Succeeded)
		})).To(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(K8sClient.Get(Ctx, runConfiguration.GetNamespacedName(), runConfiguration)).To(Succeed())
			g.Expect(runConfiguration.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
			g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
		}).Should(Succeed())
	})

	It("deletes RunSchedule owned resources that are not in the triggers", func() {
		runConfiguration := createSucceededRcWithSchedule()

		runConfiguration.Spec.Triggers.Schedules = nil
		Expect(K8sClient.Update(Ctx, runConfiguration)).To(Succeed())

		Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
			g.Expect(fetchedRc.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
			g.Expect(fetchedRc.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
		})).Should(Succeed())

		Eventually(hasNoSchedules(runConfiguration)).Should(Succeed())
	})

	It("succeeds without creating resources if dependencies are not met and records event", func() {
		runConfiguration := createSucceededRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
			runConfiguration.Spec.Triggers = pipelineshub.RandomScheduleTrigger()
			runConfiguration.Spec.Triggers.OnChange = []pipelineshub.OnChangeType{pipelineshub.OnChangeTypes.RunSpec}
			runConfiguration.Spec.Run.Parameters = []pipelineshub.Parameter{
				pipelineshub.RandomRunConfigurationRefParameter(),
			}
			return runConfiguration
		})

		Eventually(hasNoSchedules(runConfiguration)).Should(Succeed())
		Eventually(hasNoRuns(runConfiguration)).Should(Succeed())

		Eventually(func(g Gomega) {
			events := v1.EventList{}
			g.Expect(K8sClient.List(Ctx, &events, client.MatchingFields{"involvedObject.name": runConfiguration.Name})).To(Succeed())

			g.Expect(events.Items).To(ContainElement(And(
				HaveReason(EventReasons.Synced),
				HaveField("Message", ContainSubstring("Unable to resolve parameters")),
			)))
		}).Should(Succeed())
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

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = pipelineshub.PipelineIdentifier{Name: apis.RandomString(), Version: pipelineVersion}
				return runConfiguration
			})

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.Dependencies.Pipeline.Version).To(Equal(pipelineVersion))
			})).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with no version specified on the RC", func() {
		It("triggers an update of the run configuration", func() {
			pipeline := pipelineshub.RandomPipeline(Provider.GetCommonNamespacedName())
			pipeline.Spec.Framework.Name = TestFramework
			pipelineHelper := CreateSucceeded(pipeline)

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
				runConfiguration.Status.Dependencies.Pipeline.Version = pipeline.ComputeVersion()
				return runConfiguration
			})

			pipelineHelper.UpdateStable(func(pipeline *pipelineshub.Pipeline) {
				pipeline.Spec = pipelineshub.RandomPipelineSpec(Provider.GetCommonNamespacedName())
				pipeline.Spec.Framework.Name = TestFramework
			})

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.Dependencies.Pipeline.Version).To(Equal(pipeline.ComputeVersion()))
			})).Should(Succeed())
		})

		It("change trigger creates a run when the pipeline is updated", func() {
			pipeline := pipelineshub.RandomPipeline(Provider.GetCommonNamespacedName())
			pipeline.Spec.Framework.Name = TestFramework
			pipelineHelper := CreateSucceeded(pipeline)
			firstPipelineVersion := pipeline.ComputeVersion()

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
				runConfiguration.Spec.Triggers = pipelineshub.RandomOnChangeTrigger()
				return runConfiguration
			})

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(ConsistOf(HavePipelineVersion(firstPipelineVersion)))
			}).Should(Succeed())

			pipelineHelper.UpdateStable(func(pipeline *pipelineshub.Pipeline) {
				pipeline.Spec = pipelineshub.RandomPipelineSpec(Provider.GetCommonNamespacedName())
				pipeline.Spec.Framework.Name = TestFramework
			})

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(ConsistOf(HavePipelineVersion(firstPipelineVersion), HavePipelineVersion(pipeline.ComputeVersion())))
			}).Should(Succeed())
		})

		It("change trigger creates a run when the pipeline is created", func() {
			pipeline := pipelineshub.RandomPipeline(Provider.GetCommonNamespacedName())
			pipeline.Spec.Framework.Name = TestFramework
			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
				runConfiguration.Spec.Triggers = pipelineshub.RandomOnChangeTrigger()
				return runConfiguration
			})

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
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
			pipeline := pipelineshub.RandomPipeline(Provider.GetCommonNamespacedName())
			pipeline.Spec.Framework.Name = TestFramework
			fixedIdentifier := pipeline.VersionedIdentifier()

			pipelineHelper := CreateSucceeded(pipeline)

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
				runConfiguration.Spec.Run.Pipeline = fixedIdentifier
				runConfiguration.Status.Dependencies.Pipeline.Version = pipeline.ComputeVersion()
				return runConfiguration
			})

			pipelineHelper.UpdateStable(func(pipeline *pipelineshub.Pipeline) {
				pipeline.Spec = pipelineshub.RandomPipelineSpec(Provider.GetCommonNamespacedName())
				pipeline.Spec.Framework.Name = TestFramework
			})

			// To verify the absence of additional RC updates, force another update of the resource.
			// If the update is processed but the pipeline version hasn't changed,
			// given that reconciliation requests are processed in-order, we can conclude that the RC is fixed.
			newExperiment := apis.RandomString()

			runConfiguration.Spec.Run.ExperimentName = newExperiment
			Expect(K8sClient.Update(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Spec.Run.ExperimentName).To(Equal(newExperiment))
				g.Expect(fetchedRc.Status.Dependencies.Pipeline.Version).To(Equal(fixedIdentifier.Version))
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration does not exist", func() {
		It("unsets the dependency", func() {
			runConfigurationName := common.RandomNamespacedName()

			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
				runConfiguration.Spec.Run.Parameters = []pipelineshub.Parameter{
					{
						ValueFrom: &pipelineshub.ValueFrom{
							RunConfigurationRef: pipelineshub.RunConfigurationRef{
								Name:           runConfigurationName,
								OutputArtifact: apis.RandomString(),
							},
						},
					},
				}
				return runConfiguration
			})
			rcNamespacedName, err := runConfigurationName.String()
			Expect(err).NotTo(HaveOccurred())

			runConfiguration.Status.Dependencies.RunConfigurations = map[string]pipelineshub.RunReference{
				rcNamespacedName: {
					ProviderId: apis.RandomString(),
					Artifacts:  []common.Artifact{common.RandomArtifact()},
				},
			}

			Expect(K8sClient.Status().Update(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.ObservedGeneration).To(Equal(fetchedRc.Generation))
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[rcNamespacedName].ProviderId).To(BeEmpty())
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[rcNamespacedName].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A referenced RunConfiguration has succeeded but misses the outputArtifact", func() {
		It("unsets the dependency", func() {
			referencedRc := createRcWithLatestRun(pipelineshub.RunReference{
				ProviderId: apis.RandomString(),
			})
			referencedRcNamespacedName := common.NamespacedName{Name: referencedRc.Name, Namespace: referencedRc.Namespace}

			runConfiguration := pipelineshub.RandomRunConfiguration(Provider.GetCommonNamespacedName())
			runConfiguration.Spec.Run.Parameters = []pipelineshub.Parameter{
				{
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           referencedRcNamespacedName,
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}
			rcNamespacedName, err := referencedRcNamespacedName.String()
			Expect(err).NotTo(HaveOccurred())

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			oldState := runConfiguration.Status.Conditions.SynchronizationSucceeded()
			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.Conditions.SynchronizationSucceeded()).To(Equal(oldState))
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[rcNamespacedName].ProviderId).To(BeEmpty())
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[rcNamespacedName].Artifacts).To(BeEmpty())
			})).Should(Succeed())
		})
	})

	When("A RunConfiguration reference has been removed", func() {
		It("removes the dependency", func() {
			runConfiguration := createSucceededRc()

			excessDependency := apis.RandomString()
			runConfiguration.SetDependencyRuns(map[string]pipelineshub.RunReference{excessDependency: {}})

			Expect(K8sClient.Status().Update(Ctx, runConfiguration)).To(Succeed())

			oldState := runConfiguration.Status.Conditions.SynchronizationSucceeded()
			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.Conditions.SynchronizationSucceeded()).To(Equal(oldState))
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations).NotTo(HaveKey(excessDependency))
			})).Should(Succeed())
		})
	})

	When("Completing the referenced run configuration", func() {
		It("Sets the run configuration's dependency field", func() {
			referencedRc := createRcWithLatestRun(pipelineshub.RunReference{
				ProviderId: apis.RandomString(),
				Artifacts:  []common.Artifact{common.RandomArtifact()},
			})
			referencedRcNamespacedName := common.NamespacedName{Name: referencedRc.Name, Namespace: referencedRc.Namespace}
			runConfiguration := pipelineshub.RandomRunConfiguration(Provider.GetCommonNamespacedName())
			runConfiguration.Spec.Run.Parameters = []pipelineshub.Parameter{
				{
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           referencedRcNamespacedName,
							OutputArtifact: referencedRc.Status.LatestRuns.Succeeded.Artifacts[0].Name,
						},
					},
				},
			}

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			rcNamespacedName, err := referencedRcNamespacedName.String()
			Expect(err).NotTo(HaveOccurred())

			Expect(K8sClient.Get(Ctx, referencedRc.GetNamespacedName(), referencedRc)).To(Succeed())
			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.Dependencies.RunConfigurations[rcNamespacedName]).To(Equal(referencedRc.Status.LatestRuns.Succeeded))
			})).Should(Succeed())
		})

		It("RC trigger creates a run when the referenced RC has finished", func() {
			referencedRc := createRcWithLatestRun(pipelineshub.RunReference{
				ProviderId: apis.RandomString(),
			})
			referencedRcNamespacedName := common.NamespacedName{Name: referencedRc.Name, Namespace: referencedRc.Namespace}

			runConfiguration := pipelineshub.RandomRunConfiguration(Provider.GetCommonNamespacedName())
			runConfiguration.Spec.Triggers = pipelineshub.Triggers{
				RunConfigurations: []common.NamespacedName{
					referencedRcNamespacedName,
				},
			}

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			rcNamespacedName, err := referencedRcNamespacedName.String()
			Expect(err).NotTo(HaveOccurred())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(runConfiguration.Status.Dependencies.RunConfigurations[rcNamespacedName].ProviderId).To(Equal(referencedRc.Status.LatestRuns.Succeeded.ProviderId))
			})).Should(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.Triggers.RunConfigurations[rcNamespacedName].ProviderId).To(Equal(referencedRc.Status.LatestRuns.Succeeded.ProviderId))
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
			referencedRc := createRcWithLatestRun(pipelineshub.RunReference{
				ProviderId: apis.RandomString(),
			})
			referencedRcNamespacedName := common.NamespacedName{Name: referencedRc.Name, Namespace: referencedRc.Namespace}
			runConfiguration := createSucceededRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
				runConfiguration.Spec.Triggers = pipelineshub.Triggers{
					RunConfigurations: []common.NamespacedName{
						referencedRcNamespacedName,
					},
				}
				return runConfiguration
			})

			runConfiguration.Spec.Triggers.RunConfigurations = nil
			Expect(K8sClient.Update(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.Triggers.RunConfigurations).NotTo(HaveKey(referencedRcNamespacedName))
			})).Should(Succeed())
		})
	})

	When("RunSpec changes", func() {
		It("onChange trigger creates a run when run template has changed", func() {
			runConfiguration := pipelineshub.RandomRunConfiguration(Provider.GetCommonNamespacedName())
			runConfiguration.Spec.Triggers = pipelineshub.Triggers{
				OnChange: []pipelineshub.OnChangeType{
					pipelineshub.OnChangeTypes.RunSpec,
				},
			}

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(runConfiguration.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Succeeded))
			})).Should(Succeed())

			runConfiguration.Spec.Run = pipelineshub.RandomRunSpec(Provider.GetCommonNamespacedName())

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(HaveLen(1))
			}).Should(Succeed())
		})
	})

	When("setting the provider", func() {
		It("stores the provider in the status", func() {
			runConfiguration := pipelineshub.RandomRunConfiguration(Provider.GetCommonNamespacedName())
			runConfiguration.Spec.Run.Provider.Name = Provider.Name
			runConfiguration.Spec.Run.Provider.Namespace = Provider.Namespace
			runConfiguration.Spec.Triggers = pipelineshub.Triggers{}
			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.Provider).To(Equal(runConfiguration.Spec.Run.Provider))
			})).Should(Succeed())
		})

		It("passes the provider to owned resources", func() {
			runConfiguration := pipelineshub.RandomRunConfiguration(Provider.GetCommonNamespacedName())
			runConfiguration.Spec.Run.Provider.Name = Provider.Name
			runConfiguration.Spec.Triggers = pipelineshub.Triggers{
				Schedules: []pipelineshub.Schedule{pipelineshub.RandomSchedule()},
				OnChange:  []pipelineshub.OnChangeType{pipelineshub.OnChangeTypes.Pipeline},
			}
			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelineshub.RunSchedule) {
				g.Expect(ownedSchedule.Spec.Provider).To(Equal(runConfiguration.Spec.Run.Provider))
			})).Should(Succeed())
			Eventually(matchRuns(runConfiguration, func(g Gomega, ownedRun *pipelineshub.Run) {
				g.Expect(ownedRun.Spec.Provider).To(Equal(runConfiguration.Spec.Run.Provider))
			})).Should(Succeed())
		})
	})

	When("changing the provider", func() {
		It("fails the resource", func() {
			runConfiguration := createSucceededRcWithSchedule()
			runConfiguration.Spec.Run.Provider.Name = apis.RandomLowercaseString()
			Expect(K8sClient.Update(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
				g.Expect(fetchedRc.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(apis.Failed))
			})).Should(Succeed())

			Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelineshub.RunSchedule) {
				g.Expect(ownedSchedule.Spec.Provider.Name).To(Equal(Provider.Name))
				g.Expect(ownedSchedule.Spec.Provider.Namespace).To(Equal(Provider.Namespace))
			})).Should(Succeed())
		})
	})

	When("Creating an invalid run configuration", func() {
		It("errors", func() {
			runConfiguration := pipelineshub.RandomRunConfiguration(Provider.GetCommonNamespacedName())
			runConfiguration.Spec.Run.Parameters = []pipelineshub.Parameter{
				{
					Value: apis.RandomString(),
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           common.RandomNamespacedName(),
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
			runConfiguration.Spec.Run.Parameters = []pipelineshub.Parameter{
				{
					Value: apis.RandomString(),
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           common.RandomNamespacedName(),
							OutputArtifact: apis.RandomString(),
						},
					},
				},
			}

			Expect(K8sClient.Update(Ctx, runConfiguration)).To(MatchError(ContainSubstring("only one of value or valueFrom can be set")))
		})
	})

	When("A run configuration has unresolved optional parameters", func() {
		It("records events for unresolved optional parameters", func() {
			referencedRc := createRcWithLatestRun(pipelineshub.RunReference{
				ProviderId: apis.RandomString(),
				Artifacts: []common.Artifact{
					common.RandomArtifact(),
				},
			})
			referencedRcNamespacedName := common.NamespacedName{
				Name:      referencedRc.Name,
				Namespace: referencedRc.Namespace,
			}

			runConfiguration := pipelineshub.RandomRunConfiguration(Provider.GetCommonNamespacedName())
			optParamName := "optional-param"
			runConfiguration.Spec.Run.Parameters = []pipelineshub.Parameter{
				{
					Name: "working-param",
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           referencedRcNamespacedName,
							OutputArtifact: referencedRc.Status.LatestRuns.Succeeded.Artifacts[0].Name,
						},
					},
				},
				{
					Name: optParamName,
					ValueFrom: &pipelineshub.ValueFrom{
						RunConfigurationRef: pipelineshub.RunConfigurationRef{
							Name:           referencedRcNamespacedName,
							OutputArtifact: "missing-artifact",
							Optional:       true,
						},
					},
				},
			}

			rcNamespacedName, err := referencedRcNamespacedName.String()
			Expect(err).NotTo(HaveOccurred())

			Expect(K8sClient.Create(Ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, rc *pipelineshub.RunConfiguration) {
				g.Expect(rc.Status.Dependencies.RunConfigurations[rcNamespacedName]).To(Equal(referencedRc.Status.LatestRuns.Succeeded))
			})).Should(Succeed())

			Eventually(func(g Gomega) {
				events := v1.EventList{}
				g.Expect(K8sClient.List(Ctx, &events, client.MatchingFields{"involvedObject.name": runConfiguration.Name})).To(Succeed())

				g.Expect(events.Items).To(ContainElement(And(
					HaveReason(EventReasons.Synced),
					HaveField("Message", ContainSubstring(fmt.Sprintf("Unable to resolve parameter '%s', but skipping as it is marked as optional.", optParamName))),
				)))
			}).Should(Succeed())
		})
	})
})

func matchRuns(runConfiguration *pipelineshub.RunConfiguration, matcher func(Gomega, *pipelineshub.Run)) func(Gomega) {
	return func(g Gomega) {
		ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedRuns).NotTo(BeEmpty())
		for _, ownedRun := range ownedRuns {
			matcher(g, &ownedRun)
		}
	}
}

func hasNoRuns(runConfiguration *pipelineshub.RunConfiguration) func(Gomega) {
	return func(g Gomega) {
		ownedRuns, err := findOwnedRuns(Ctx, K8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedRuns).To(BeEmpty())
	}
}

func hasNoSchedules(runConfiguration *pipelineshub.RunConfiguration) func(Gomega) {
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
	run, isRun := actual.(pipelineshub.Run)
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
