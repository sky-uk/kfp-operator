//go:build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	pipelinesv1 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Context("aggregateState", func() {
	DescribeTable("calculates based on sub states", func(subStates []apis.SynchronizationState, expectedState apis.SynchronizationState, expectedMessage string) {
		runSchedules := make([]pipelinesv1.RunSchedule, len(subStates))
		for i, state := range subStates {
			runSchedules[i] = pipelinesv1.RunSchedule{Status: pipelinesv1.Status{SynchronizationState: state, Conditions: []metav1.Condition{{Type: pipelinesv1.ConditionTypes.SynchronizationSucceeded, Message: string(state)}}}}
		}

		state, message := aggregateState(runSchedules)
		Expect(state).To(Equal(expectedState))
		Expect(message).To(Equal(expectedMessage))
	},
		Entry(nil, []apis.SynchronizationState{}, apis.Succeeded, ""),
		Entry(nil, []apis.SynchronizationState{apis.Failed, apis.Succeeded}, apis.Failed, "Failed"),
		Entry(nil, []apis.SynchronizationState{apis.Updating, apis.Failed}, apis.Updating, ""),
		Entry(nil, []apis.SynchronizationState{apis.Deleting, apis.Failed}, apis.Updating, ""),
		Entry(nil, []apis.SynchronizationState{apis.Deleted, apis.Failed}, apis.Updating, ""),
		Entry(nil, []apis.SynchronizationState{"", apis.Failed}, apis.Updating, ""),
		Entry(nil, []apis.SynchronizationState{apis.Succeeded}, apis.Succeeded, ""),
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
			Expect(schedule.Spec.Pipeline.Version).To(Equal(runConfiguration.Status.Dependencies.Pipeline.Version))
			Expect(schedule.Spec.RuntimeParameters).To(Equal(pipelines.Map(runConfiguration.Spec.Run.RuntimeParameters, func(rp pipelinesv1.RuntimeParameter) apis.NamedValue {
				return apis.NamedValue{Name: rp.Name, Value: rp.Value}
			})))
			Expect(schedule.Spec.ExperimentName).To(Equal(runConfiguration.Spec.Run.ExperimentName))
			Expect(schedule.Spec.Schedule).To(Equal(runConfiguration.Spec.Triggers.Schedules[i]))
			Expect(metav1.IsControlledBy(&schedule, runConfiguration)).To(BeTrue())
			Expect(schedule.GetAnnotations()[apis.ResourceAnnotations.Provider]).To(Equal(provider))
		}
	})
})

var _ = Context("updateRcTriggers", PropertyBased, func() {
	It("sets the pipeline trigger status", func() {
		runConfiguration := pipelinesv1.RandomRunConfiguration()
		runConfiguration.Spec.Triggers.OnChange = []pipelinesv1.OnChangeType{
			pipelinesv1.OnChangeTypes.Pipeline,
		}
		runConfiguration.Status.Dependencies.Pipeline.Version = apis.RandomString()
		rcr := RunConfigurationReconciler{}
		Expect(rcr.updateRcTriggers(*runConfiguration).Triggers.Pipeline.Version).To(Equal(runConfiguration.Status.Dependencies.Pipeline.Version))
	})

	It("sets the runSpec trigger status", func() {
		runConfiguration := pipelinesv1.RandomRunConfiguration()
		runConfiguration.Spec.Triggers.OnChange = []pipelinesv1.OnChangeType{
			pipelinesv1.OnChangeTypes.RunSpec,
		}
		rcr := RunConfigurationReconciler{}
		Expect(rcr.updateRcTriggers(*runConfiguration).Triggers.RunSpec.Version).To(Equal(runConfiguration.Spec.Run.ComputeVersion()))
	})

	It("sets the runConfigurations trigger status", func() {
		runConfiguration := pipelinesv1.RandomRunConfiguration()
		runConfiguration.Spec.Triggers.RunConfigurations = apis.RandomList(apis.RandomString)
		runConfiguration.Status.Dependencies.RunConfigurations = make(map[string]pipelinesv1.RunReference)
		for _, rc := range runConfiguration.Spec.Triggers.RunConfigurations {
			runConfiguration.Status.Dependencies.RunConfigurations[rc] = pipelinesv1.RunReference{
				ProviderId: apis.RandomString(),
			}
		}
		rcr := RunConfigurationReconciler{}
		updatedStatus := rcr.updateRcTriggers(*runConfiguration)
		for _, rc := range runConfiguration.Spec.Triggers.RunConfigurations {
			Expect(updatedStatus.Triggers.RunConfigurations[rc].ProviderId).To(Equal(runConfiguration.Status.Dependencies.RunConfigurations[rc].ProviderId))
		}
		Expect(updatedStatus.Triggers.RunConfigurations).To(HaveLen(len(runConfiguration.Spec.Triggers.RunConfigurations)))
	})

	It("retains other fields", func() {
		runConfiguration := pipelinesv1.RandomRunConfiguration()
		rcr := RunConfigurationReconciler{}
		updatedStatus := rcr.updateRcTriggers(*runConfiguration)
		updatedStatus.Triggers = runConfiguration.Status.Triggers
		updatedStatus.Triggers.Pipeline.Version = runConfiguration.Status.Triggers.Pipeline.Version
		Expect(updatedStatus).To(Equal(runConfiguration.Status))
	})
})
