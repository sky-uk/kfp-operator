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

var _ = Context("aggregateState", func() {
	DescribeTable("calculates based on sub states", func(subStates []apis.SynchronizationState, expected apis.SynchronizationState) {
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
})

var _ = Context("constructRunSchedulesForTriggers", PropertyBased, func() {
	Expect(pipelinesv1.AddToScheme(scheme.Scheme)).To(Succeed())
	rcr := RunConfigurationReconciler{
		Scheme: scheme.Scheme,
	}

	It("sets all spec fields", func() {
		runConfiguration := pipelinesv1.RandomRunConfiguration()
		runConfiguration.Spec.Triggers = pipelinesv1.Triggers{Schedules: apis.RandomList(apis.RandomString)}
		provider := apis.RandomString()

		runSchedules, err := rcr.constructRunSchedulesForTriggers(provider, runConfiguration)
		Expect(err).NotTo(HaveOccurred())

		for i, schedule := range runSchedules {
			Expect(schedule.Namespace).To(Equal(runConfiguration.Namespace))
			Expect(schedule.Spec.Pipeline.Name).To(Equal(runConfiguration.Spec.Run.Pipeline.Name))
			Expect(schedule.Spec.Pipeline.Version).To(Equal(runConfiguration.Status.ObservedPipelineVersion))
			Expect(schedule.Spec.RuntimeParameters).To(Equal(runConfiguration.Spec.Run.RuntimeParameters))
			Expect(schedule.Spec.ExperimentName).To(Equal(runConfiguration.Spec.Run.ExperimentName))
			Expect(schedule.Spec.Schedule).To(Equal(runConfiguration.Spec.Triggers.Schedules[i]))
			Expect(metav1.IsControlledBy(&schedule, runConfiguration)).To(BeTrue())
			Expect(schedule.GetAnnotations()[apis.ResourceAnnotations.Provider]).To(Equal(provider))
		}
	})
})
