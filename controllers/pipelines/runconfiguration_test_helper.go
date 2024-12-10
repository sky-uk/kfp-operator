//go:build decoupled

package pipelines

import (
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
)

func createSucceededRcWithSchedule() *pipelinesv1.RunConfiguration {
	runConfiguration := createStableRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
		runConfiguration.Spec.Triggers = pipelinesv1.RandomScheduleTrigger()
		return runConfiguration
	}, apis.Updating)

	Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelinesv1.RunSchedule) {
		g.Expect(ownedSchedule.Status.SynchronizationState).To(Equal(apis.Creating))
	})).Should(Succeed())

	Expect(updateOwnedSchedules(runConfiguration, func(ownedSchedule *pipelinesv1.RunSchedule) {
		ownedSchedule.Status.SynchronizationState = apis.Succeeded
	})).To(Succeed())

	Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
		g.Expect(runConfiguration.Status.SynchronizationState).To(Equal(apis.Succeeded))
	})).Should(Succeed())

	return runConfiguration
}

func createSucceededRc() *pipelinesv1.RunConfiguration {
	return createStableRcWith(func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
		return runConfiguration
	}, apis.Succeeded)
}

func createSucceededRcWith(modifyRc func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration {
	return createStableRcWith(modifyRc, apis.Succeeded)
}

func createStableRcWith(modifyRc func(runConfiguration *pipelinesv1.RunConfiguration) *pipelinesv1.RunConfiguration, synchronizationState apis.SynchronizationState) *pipelinesv1.RunConfiguration {
	runConfiguration := pipelinesv1.RandomRunConfiguration(Provider.Name)
	runConfiguration.Spec.Run.RuntimeParameters = []pipelinesv1.RuntimeParameter{}
	runConfiguration.Spec.Triggers = pipelinesv1.Triggers{}
	modifiedRc := modifyRc(runConfiguration)
	Expect(K8sClient.Create(Ctx, modifiedRc)).To(Succeed())

	Eventually(matchRunConfiguration(modifiedRc, func(g Gomega, fetchedRc *pipelinesv1.RunConfiguration) {
		g.Expect(fetchedRc.Status.ObservedGeneration).To(Equal(modifiedRc.Generation))
		g.Expect(fetchedRc.Status.SynchronizationState).To(Equal(synchronizationState))
		g.Expect(fetchedRc.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(synchronizationState))
		modifiedRc = fetchedRc
	})).Should(Succeed())

	return modifiedRc
}

func createRcWithLatestRun(succeeded pipelinesv1.RunReference) *pipelinesv1.RunConfiguration {
	referencedRc := createSucceededRc()
	referencedRc.Status.LatestRuns.Succeeded = succeeded
	Expect(K8sClient.Status().Update(Ctx, referencedRc)).To(Succeed())

	return referencedRc
}

func matchRunConfiguration(runConfiguration *pipelinesv1.RunConfiguration, matcher func(Gomega, *pipelinesv1.RunConfiguration)) func(Gomega) {
	return func(g Gomega) {
		g.Expect(K8sClient.Get(Ctx, runConfiguration.GetNamespacedName(), runConfiguration)).To(Succeed())
		matcher(g, runConfiguration)
	}
}

func updateOwnedSchedules(runConfiguration *pipelinesv1.RunConfiguration, updateFn func(schedule *pipelinesv1.RunSchedule)) error {
	ownedSchedules, err := findOwnedRunSchedules(Ctx, K8sClient, runConfiguration)
	if err != nil {
		return err
	}

	for _, ownedSchedule := range ownedSchedules {
		updateFn(&ownedSchedule)
		Expect(K8sClient.Status().Update(Ctx, &ownedSchedule)).To(Succeed())
	}

	return nil
}

func matchSchedules(runConfiguration *pipelinesv1.RunConfiguration, matcher func(Gomega, *pipelinesv1.RunSchedule)) func(Gomega) {
	return func(g Gomega) {
		ownedSchedules, err := findOwnedRunSchedules(Ctx, K8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedSchedules).NotTo(BeEmpty())
		for _, ownedSchedule := range ownedSchedules {
			matcher(g, &ownedSchedule)
		}
	}
}
