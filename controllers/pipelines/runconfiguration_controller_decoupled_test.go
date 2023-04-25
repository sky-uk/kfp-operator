//go:build decoupled
// +build decoupled

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("RunConfiguration controller k8s integration", Serial, func() {
	It("creates RunSchedule owned resources that are in the triggers", func() {
		runConfiguration := pipelinesv1.RandomRunConfiguration()
		runConfiguration.Spec.Triggers = []pipelinesv1.Trigger{pipelinesv1.RandomCronTrigger()}
		Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

		Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
			g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Updating))
			g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
		})).Should(Succeed())

		Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelinesv1.RunSchedule) {
			g.Expect(ownedSchedule.Spec.Pipeline).To(Equal(runConfiguration.Spec.Run.Pipeline))
			g.Expect(ownedSchedule.Spec.RuntimeParameters).To(Equal(runConfiguration.Spec.Run.RuntimeParameters))
			g.Expect(ownedSchedule.Spec.Schedule).To(Equal(runConfiguration.Spec.Triggers[0].CronExpression))
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

		runConfiguration.Spec.Triggers = []pipelinesv1.Trigger{}
		Expect(k8sClient.Update(ctx, runConfiguration)).To(Succeed())

		Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, configuration *pipelinesv1.RunConfiguration) {
			g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Succeeded))
			g.Expect(runConfiguration.Status.ObservedGeneration).To(Equal(runConfiguration.GetGeneration()))
		})).Should(Succeed())

		Eventually(hasNoSchedules(runConfiguration)).Should(Succeed())
	})

	It("cascades deletes when the RunConfiguration is deleted", func() {
		Skip("See https://github.com/kubernetes-sigs/controller-runtime/issues/1459. Keep test for documentation")
		runConfiguration := createSucceededRcWithSchedule()

		Expect(k8sClient.Delete(ctx, runConfiguration)).To(Succeed())

		Eventually(hasNoSchedules(runConfiguration)).Should(Succeed())
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

		It("creates a run if there is a change trigger", func() {
			pipeline := pipelinesv1.RandomPipeline()
			pipelineHelper := CreateSucceeded(pipeline)

			runConfiguration := pipelinesv1.RandomRunConfiguration()
			runConfiguration.Spec.Run.Pipeline = pipeline.UnversionedIdentifier()
			runConfiguration.Status.ObservedPipelineVersion = pipeline.ComputeVersion()

			runConfiguration.Spec.Triggers = []pipelinesv1.Trigger{{Type: pipelinesv1.TriggerTypes.Change}}

			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())

			pipelineHelper.UpdateStable(func(pipeline *pipelinesv1.Pipeline) {
				pipeline.Spec = pipelinesv1.RandomPipelineSpec()
			})

			Eventually(func(g Gomega) {
				ownedRuns, err := findOwnedRuns(runConfiguration)
				Expect(err).NotTo(HaveOccurred())
				Expect(ownedRuns).To(HaveLen(1))
				Expect(ownedRuns[0].Spec.Pipeline).To(Equal(pipeline.ComputeVersion()))
			}).Should(Succeed())
		})
	})

	When("Updating the referenced pipeline with a fixed version specified on the RC", func() {
		It("does not trigger an update of the run configuration", func() {
			pipeline := pipelinesv1.RandomPipeline()
			fixedIdentifier := pipeline.VersionedIdentifier()

			runConfiguration := pipelinesv1.RandomRunConfiguration()

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

	When("Validation fails", func() {
		It("fails creates", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration()

			runConfiguration.Spec.Triggers = []pipelinesv1.Trigger{
				{Type: "not a type"},
			}

			Expect(k8sClient.Create(ctx, runConfiguration)).To(MatchError(
				ContainSubstring("is invalid: spec.triggers[0].type"),
			))
		})

		It("fails updates", func() {
			runConfiguration := pipelinesv1.RandomRunConfiguration()
			Expect(k8sClient.Create(ctx, runConfiguration)).To(Succeed())
			Expect(k8sClient.Get(ctx, runConfiguration.GetNamespacedName(), runConfiguration)).To(Succeed())

			runConfiguration.Spec.Triggers = []pipelinesv1.Trigger{
				{Type: "not a type"},
			}

			Expect(k8sClient.Update(ctx, runConfiguration)).To(MatchError(
				ContainSubstring("is invalid: spec.triggers[0].type"),
			))
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

func hasNoSchedules(runConfiguration *pipelinesv1.RunConfiguration) func(Gomega) {
	return func(g Gomega) {
		ownedSchedules, err := findOwnedRunSchedules(ctx, k8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedSchedules).To(BeEmpty())
	}
}

func createSucceededRcWithSchedule() *pipelinesv1.RunConfiguration {
	runConfiguration := pipelinesv1.RandomRunConfiguration()
	runConfiguration.Spec.Triggers = append(runConfiguration.Spec.Triggers, pipelinesv1.RandomCronTrigger())
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

func findOwnedRuns(runConfiguration *pipelinesv1.RunConfiguration) ([]pipelinesv1.Run, error) {
	ownedRunsList := &pipelinesv1.RunList{}
	if err := k8sClient.List(ctx, ownedRunsList, client.InNamespace(runConfiguration.Namespace)); err != nil {
		return nil, err
	}

	var ownedRuns []pipelinesv1.Run
	for _, run := range ownedRunsList.Items {
		if metav1.IsControlledBy(&run, runConfiguration) {
			ownedRuns = append(ownedRuns, run)
		}
	}

	return ownedRuns, nil
}
