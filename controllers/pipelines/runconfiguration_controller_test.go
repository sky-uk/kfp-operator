//go:build unit
// +build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Context("RunConfigurationController", func() {
	DescribeTable("aggregateState", func(subStates []apis.SynchronizationState, expected apis.SynchronizationState) {
		runSchedules := make([]pipelinesv1.RunSchedule, len(subStates))
		for i, state := range subStates {
			runSchedules[i] = pipelinesv1.RunSchedule{Status: pipelinesv1.Status{SynchronizationState: state}}
		}

		Expect(aggregateState(runSchedules)).To(Equal(expected))
	},
		Entry(nil, []apis.SynchronizationState{}, apis.Succeeded),
		Entry(nil, []apis.SynchronizationState{apis.Failed, apis.Succeeded}, apis.Failed),
		Entry(nil, []apis.SynchronizationState{apis.Updating, apis.Failed}, apis.Updating),
		Entry(nil, []apis.SynchronizationState{apis.Deleting, apis.Failed}, apis.Updating),
		Entry(nil, []apis.SynchronizationState{apis.Deleted, apis.Failed}, apis.Updating),
		Entry(nil, []apis.SynchronizationState{"", apis.Failed}, apis.Updating),
		Entry(nil, []apis.SynchronizationState{apis.Succeeded}, apis.Succeeded),
	)

	It("constructRunSchedulesForTriggers", func() {
		runConfiguration := pipelinesv1.RandomRunConfiguration()
		runConfiguration.Spec.Triggers = []pipelinesv1.Trigger{pipelinesv1.RandomTrigger(), pipelinesv1.RandomTrigger(), pipelinesv1.RandomTrigger()}

		Expect(pipelinesv1.AddToScheme(scheme.Scheme)).To(Succeed())
		runSchedules, err := constructRunSchedulesForTriggers(runConfiguration, scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		for i, schedule := range runSchedules {
			Expect(schedule.Namespace).To(Equal(runConfiguration.Namespace))
			Expect(schedule.Spec.Pipeline.Name).To(Equal(runConfiguration.Spec.Pipeline.Name))
			Expect(schedule.Spec.Pipeline.Version).To(Equal(runConfiguration.Status.ObservedPipelineVersion))
			Expect(schedule.Spec.RuntimeParameters).To(Equal(runConfiguration.Spec.RuntimeParameters))
			Expect(schedule.Spec.ExperimentName).To(Equal(runConfiguration.Spec.ExperimentName))
			Expect(schedule.Spec.Schedule).To(Equal(runConfiguration.Spec.Triggers[i].CronExpression))
			Expect(metav1.IsControlledBy(&schedule, runConfiguration)).To(BeTrue())
		}
	})
})
