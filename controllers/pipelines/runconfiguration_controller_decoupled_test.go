//go:build decoupled
// +build decoupled

package pipelines

import (
	"fmt"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("RunConfiguration controller k8s integration", Serial, func() {
	It("creates RunSchedule owned resources that are in the triggers", func() {
		runConfiguration := pipelinesv1.RandomRunConfiguration()
		runConfiguration.Spec.Triggers = pipelinesv1.RandomScheduleTrigger()
		Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

		Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
			g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Updating))
			g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
		})).Should(Succeed())

		Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelinesv1.RunSchedule) {
			g.Expect(ownedSchedule.Spec.Pipeline).To(Equal(runConfiguration.Spec.Run.Pipeline))
			g.Expect(ownedSchedule.Spec.RuntimeParameters).To(Equal(runConfiguration.Spec.Run.RuntimeParameters))
			g.Expect(ownedSchedule.Spec.Schedule).To(Equal(runConfiguration.Spec.Triggers.Schedules[0]))
			g.Expect(ownedSchedule.Status.SynchronizationState).To(Equal(apis.Creating))
		})).Should(Succeed())

		Expect(updateOwnedSchedules(runConfiguration, func(ownedSchedule *pipelinesv1.RunSchedule) {
			ownedSchedule.Status.SynchronizationState = apis.Succeeded
		})).To(Succeed())

		Eventually(func(g Gomega) {
			g.Expect(k8sClient.Get(ctx, runConfiguration.GetNamespacedName(), runConfiguration)).To(Succeed())
			g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Succeeded))
			g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
		}).Should(Succeed())
	})

	It("deletes RunSchedule owned resources that are not in the triggers", func() {
		runConfiguration := createSucceededRcWithSchedule()

		runConfiguration.Spec.Triggers.Schedules = nil
		Expect(k8sClient.Update(ctx, runConfiguration)).To(Succeed())

		Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
			g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Updating))
			g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
		})).Should(Succeed())

		Eventually(hasNoSchedules(runConfiguration)).Should(Succeed())
	})

	When("Deleted", func() {

		It("cascades deletes", func() {
			Skip("See https://github.com/kubernetes-sigs/controller-runtime/issues/1459. Keep test for documentation")
			runConfiguration := createSucceededRcWithSchedule()

			Expect(k8sClient.Delete(ctx, runConfiguration)).To(Succeed())

			Eventually(hasNoSchedules(runConfiguration)).Should(Succeed())
		})
	})

	When("Migrating from v1alpha4", func() {
		It("Releases previously acquired resources", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration()
			runConfiguration.Spec.Triggers = pipelinesv1.Triggers{}
			controllerutil.AddFinalizer(runConfiguration, finalizerName)
			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

			Expect(k8sClient.Delete(ctx, runConfiguration)).To(Succeed())

			Eventually(func(g Gomega) {
				g.Expect(k8sClient.Get(ctx, runConfiguration.GetNamespacedName(), runConfiguration)).NotTo(Succeed())
			}).Should(Succeed())
		})
	})

	When("Creating an RC with a fixed pipeline version", func() {
		It("sets the ObservedPipelineVersion to the fixed version", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration()
			pipelineVersion := apis.RandomString()
			runConfiguration.Spec.Run.Pipeline = pipelinesv1.PipelineIdentifier{Name: apis.RandomString(), Version: pipelineVersion}

			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(pipelineVersion))
			})).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with no version specified on the RC", func() {
		It("triggers an update of the run configuration", func() {
			pipeline := pipelinesv1.RandomPipeline()
			pipelineHelper := CreateSucceeded(pipeline)

			runConfiguration := pipelinesv1.RandomRunConfiguration()
			runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
			runConfiguration.Status.ObservedPipelineVersion = pipeline.ComputeVersion()

			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

			pipelineHelper.UpdateStable(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec()
			})

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(pipeline.ComputeVersion()))
			})).Should(Succeed())
		})

		It("change trigger creates a run when the pipeline is updated", func() {
			pipeline := pipelinesv1.RandomPipeline()
			pipelineHelper := CreateSucceeded(pipeline)

			runConfiguration := pipelinesv1.RandomRunConfiguration()
			runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
			runConfiguration.Spec.Triggers = pipelinesv1.RandomOnChangeTrigger()

			firstPipelineVersion := pipeline.ComputeVersion()

			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())
			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(ctx, k8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(ConsistOf(HavePipelineVersion(firstPipelineVersion)))
			}).Should(Succeed())

			pipelineHelper.UpdateStable(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec()
			})

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(ctx, k8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(ConsistOf(HavePipelineVersion(firstPipelineVersion), HavePipelineVersion(pipeline.ComputeVersion())))
			}).Should(Succeed())
		})

		It("change trigger creates a run when the pipeline is created", func() {
			pipeline := pipelinesv1.RandomPipeline()

			runConfiguration := pipelinesv1.RandomRunConfiguration()
			runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
			runConfiguration.Spec.Triggers = pipelinesv1.RandomOnChangeTrigger()

			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())
			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
			}))
			ownedRuns, err := findOwnedRuns(ctx, k8sClient, runConfiguration)
			Expect(err).NotTo(HaveOccurred())
			Expect(ownedRuns).To(BeEmpty())

			CreateSucceeded(pipeline)

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(ctx, k8sClient, runConfiguration)
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(ownedRuns).To(ConsistOf(HavePipelineVersion(pipeline.ComputeVersion())))
			}).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with a fixed version specified on the RC", func() {
		It("does not trigger an update of the run configuration", func() {
			pipeline := pipelinesv1.RandomPipeline()
			fixedIdentifier := pipeline.VersionedIdentifier()

			runConfiguration := pipelinesv1.RandomRunConfiguration()
			runConfiguration.Spec.Triggers = pipelinesv1.Triggers{}

			runConfiguration.Spec.Run.Pipeline = fixedIdentifier
			runConfiguration.Status.ObservedPipelineVersion = pipeline.ComputeVersion()

			pipelineHelper := CreateSucceeded(pipeline)

			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

			pipelineHelper.UpdateStable(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec()
			})

			// To verify the absence of additional RC updates, force another update of the resource.
			// If the update is processed but the pipeline version hasn't changed,
			// given that reconciliation requests are processed in-order, we can conclude that the RC is fixed.
			newExperiment := apis.RandomString()
			Expect(k8sClient.Get(ctx, runConfiguration.GetNamespacedName(), runConfiguration)).To(Succeed())
			runConfiguration.Spec.Run.ExperimentName = newExperiment
			Expect(k8sClient.Update(ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Spec.Run.ExperimentName).To(Equal(newExperiment))
				g.Expect(runConfiguration.Status.ObservedPipelineVersion).To(Equal(fixedIdentifier.Version))
			})).Should(Succeed())
		})
	})

	When("setting the provider", func() {
		It("stores the provider in the status", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration()
			runConfiguration.Spec.Triggers = pipelinesv1.Triggers{}
			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.Provider).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())
		})

		It("passes the provider to owned resources", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration()
			runConfiguration.Spec.Triggers = pipelinesv1.Triggers{
				Schedules: []string{apis.RandomString()},
				OnChange:  []pipelinesv1.OnChangeType{pipelinesv1.OnChangeTypes.Pipeline},
			}
			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

			Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelinesv1.RunSchedule) {
				g.Expect(ownedSchedule.GetAnnotations()[apis.ResourceAnnotations.Provider]).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())
			Eventually(matchRuns(runConfiguration, func(g Gomega, ownedRun *pipelinesv1.Run) {
				g.Expect(ownedRun.GetAnnotations()[apis.ResourceAnnotations.Provider]).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())
		})
	})

	When("changing the provider", func() {
		It("fails the resource", func() {
			runConfiguration := createSucceededRcWithSchedule()

			metav1.SetMetaDataAnnotation(&runConfiguration.ObjectMeta, apis.ResourceAnnotations.Provider, apis.RandomString())
			Expect(k8sClient.Update(ctx, runConfiguration)).To(Succeed())

			Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
				g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Failed))
			})).Should(Succeed())

			Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelinesv1.RunSchedule) {
				g.Expect(ownedSchedule.GetAnnotations()[apis.ResourceAnnotations.Provider]).To(Equal(testConfig.DefaultProvider))
			})).Should(Succeed())
		})
	})
})

func updateOwnedSchedules(runConfiguration *pipelinesv1.RunConfiguration, updateFn func(schedule *pipelinesv1.RunSchedule)) error {
	ownedSchedules, err := findOwnedRunSchedules(ctx, k8sClient, runConfiguration)
	if err != nil {
		return err
	}

	for _, ownedSchedule := range ownedSchedules {
		updateFn(&ownedSchedule)
		Expect(k8sClient.Status().Update(ctx, &ownedSchedule)).To(Succeed())
	}

	return nil
}

func matchRunConfiguration(runConfiguration *pipelinesv1.RunConfiguration, matcher func(Gomega, *pipelinesv1.RunConfiguration)) func(Gomega) {
	return func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, runConfiguration.GetNamespacedName(), runConfiguration)).To(Succeed())
		matcher(g, runConfiguration)
	}
}

func matchSchedules(runConfiguration *pipelinesv1.RunConfiguration, matcher func(Gomega, *pipelinesv1.RunSchedule)) func(Gomega) {
	return func(g Gomega) {
		ownedSchedules, err := findOwnedRunSchedules(ctx, k8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedSchedules).NotTo(BeEmpty())
		for _, ownedSchedule := range ownedSchedules {
			matcher(g, &ownedSchedule)
		}
	}
}

func matchRuns(runConfiguration *pipelinesv1.RunConfiguration, matcher func(Gomega, *pipelinesv1.Run)) func(Gomega) {
	return func(g Gomega) {
		ownedRuns, err := findOwnedRuns(ctx, k8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedRuns).NotTo(BeEmpty())
		for _, ownedRun := range ownedRuns {
			matcher(g, &ownedRun)
		}
	}
}

func hasNoSchedules(runConfiguration *pipelinesv1.RunConfiguration) func(Gomega) {
	return func(g Gomega) {
		ownedSchedules, err := findOwnedRunSchedules(ctx, k8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedSchedules).To(BeEmpty())
	}
}

func createSucceededRcWithSchedule() *pipelinesv1.RunConfiguration {
	runConfiguration := pipelinesv1.RandomRunConfiguration()
	runConfiguration.Spec.Triggers = pipelinesv1.RandomScheduleTrigger()
	Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

	Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelinesv1.RunSchedule) {
		g.Expect(ownedSchedule.Status.SynchronizationState).To(Equal(apis.Creating))
	})).Should(Succeed())

	Expect(updateOwnedSchedules(runConfiguration, func(ownedSchedule *pipelinesv1.RunSchedule) {
		ownedSchedule.Status.SynchronizationState = apis.Succeeded
	})).To(Succeed())

	Eventually(func(g Gomega) {
		g.Expect(k8sClient.Get(ctx, runConfiguration.GetNamespacedName(), runConfiguration)).To(Succeed())
		g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Succeeded))
	}).Should(Succeed())

	return runConfiguration
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
