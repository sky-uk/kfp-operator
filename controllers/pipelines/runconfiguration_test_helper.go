//go:build decoupled

package pipelines

import (
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	. "github.com/sky-uk/kfp-operator/controllers/pipelines/internal/testutil"
)

func createSucceededRcWithSchedule() *pipelineshub.RunConfiguration {
	runConfiguration := createStableRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
		runConfiguration.Spec.Triggers = pipelineshub.RandomScheduleTrigger()
		return runConfiguration
	}, apis.Updating)

	Eventually(matchSchedules(runConfiguration, func(g Gomega, ownedSchedule *pipelineshub.RunSchedule) {
		g.Expect(ownedSchedule.Status.Conditions.SynchronizationSucceeded().Reason).To(Equal(string(apis.Creating)))
	})).Should(Succeed())

	Expect(updateOwnedSchedules(runConfiguration, func(ownedSchedule *pipelineshub.RunSchedule) {
		ownedSchedule.Status.Conditions = ownedSchedule.Status.Conditions.SetReasonForSyncState(apis.Succeeded)
	})).To(Succeed())

	Eventually(matchRunConfiguration(runConfiguration, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
		g.Expect(runConfiguration.Status.Conditions.SynchronizationSucceeded().Reason).To(Equal(string(apis.Succeeded)))
	})).Should(Succeed())

	return runConfiguration
}

func createSucceededRc() *pipelineshub.RunConfiguration {
	return createStableRcWith(func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
		return runConfiguration
	}, apis.Succeeded)
}

func createSucceededRcWith(modifyRc func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration {
	return createStableRcWith(modifyRc, apis.Succeeded)
}

func createStableRcWith(modifyRc func(runConfiguration *pipelineshub.RunConfiguration) *pipelineshub.RunConfiguration, synchronizationState apis.SynchronizationState) *pipelineshub.RunConfiguration {
	runConfiguration := pipelineshub.RandomRunConfiguration(
		common.NamespacedName{
			Name:      Provider.Name,
			Namespace: Provider.Namespace,
		},
	)
	runConfiguration.Spec.Run.Parameters = []pipelineshub.Parameter{}
	runConfiguration.Spec.Triggers = pipelineshub.Triggers{}
	modifiedRc := modifyRc(runConfiguration)
	Expect(K8sClient.Create(Ctx, modifiedRc)).To(Succeed())

	Eventually(matchRunConfiguration(modifiedRc, func(g Gomega, fetchedRc *pipelineshub.RunConfiguration) {
		g.Expect(fetchedRc.Status.ObservedGeneration).To(Equal(modifiedRc.Generation))
		g.Expect(fetchedRc.Status.Conditions.SynchronizationSucceeded().Reason).To(BeEquivalentTo(synchronizationState))
		modifiedRc = fetchedRc
	})).Should(Succeed())

	return modifiedRc
}

func createRcWithLatestRun(succeeded pipelineshub.RunReference) *pipelineshub.RunConfiguration {
	referencedRc := createSucceededRc()
	referencedRc.Status.LatestRuns.Succeeded = succeeded
	Expect(K8sClient.Status().Update(Ctx, referencedRc)).To(Succeed())

	return referencedRc
}

func matchRunConfiguration(runConfiguration *pipelineshub.RunConfiguration, matcher func(Gomega, *pipelineshub.RunConfiguration)) func(Gomega) {
	return func(g Gomega) {
		g.Expect(K8sClient.Get(Ctx, runConfiguration.GetNamespacedName(), runConfiguration)).To(Succeed())
		matcher(g, runConfiguration)
	}
}

func updateOwnedSchedules(runConfiguration *pipelineshub.RunConfiguration, updateFn func(schedule *pipelineshub.RunSchedule)) error {
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

func matchSchedules(runConfiguration *pipelineshub.RunConfiguration, matcher func(Gomega, *pipelineshub.RunSchedule)) func(Gomega) {
	return func(g Gomega) {
		ownedSchedules, err := findOwnedRunSchedules(Ctx, K8sClient, runConfiguration)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(ownedSchedules).NotTo(BeEmpty())
		for _, ownedSchedule := range ownedSchedules {
			matcher(g, &ownedSchedule)
		}
	}
}
