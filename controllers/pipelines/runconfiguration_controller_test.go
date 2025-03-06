//go:build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowfactory"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Context("aggregateState", func() {
	DescribeTable("calculates based on sub states", func(subStates []apis.SynchronizationState, expectedState apis.SynchronizationState, expectedMessage string) {
		runSchedules := make([]pipelineshub.RunSchedule, len(subStates))
		for i, state := range subStates {
			runSchedules[i] = pipelineshub.RunSchedule{
				Status: pipelineshub.Status{
					SynchronizationState: state,
					Conditions: []metav1.Condition{
						{
							Type:    pipelineshub.ConditionTypes.SynchronizationSucceeded,
							Message: string(state),
						},
					},
				},
			}
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

var _ = Context("constructRunForRunConfiguration", PropertyBased, func() {
	Expect(pipelineshub.AddToScheme(scheme.Scheme)).To(Succeed())
	rcr := RunConfigurationReconciler{
		Scheme: scheme.Scheme,
	}

	It("propagates the runconfiguration's name", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(apis.RandomLowercaseString())
		runConfiguration.Spec.Triggers = pipelineshub.Triggers{Schedules: apis.RandomList(pipelineshub.RandomSchedule)}

		run, err := rcr.constructRunForRunConfiguration(runConfiguration)
		Expect(err).NotTo(HaveOccurred())

		Expect(run.GetLabels()[workflowfactory.RunConfigurationConstants.RunConfigurationNameLabelKey]).To(Equal(runConfiguration.GetName()))
	})
})

var _ = Context("constructRunSchedulesForTriggers", PropertyBased, func() {
	Expect(pipelineshub.AddToScheme(scheme.Scheme)).To(Succeed())
	rcr := RunConfigurationReconciler{
		Scheme: scheme.Scheme,
	}

	It("sets all spec fields", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(apis.RandomLowercaseString())
		runConfiguration.Spec.Triggers = pipelineshub.Triggers{Schedules: apis.RandomList(pipelineshub.RandomSchedule)}
		resolvedParameters := pipelines.Map(runConfiguration.Spec.Run.RuntimeParameters, func(rp pipelineshub.RuntimeParameter) apis.NamedValue {
			return apis.NamedValue{Name: rp.Name, Value: rp.Value}
		})

		runSchedules, err := rcr.constructRunSchedulesForTriggers(runConfiguration, resolvedParameters)
		Expect(err).NotTo(HaveOccurred())

		for i, schedule := range runSchedules {
			Expect(schedule.Namespace).To(Equal(runConfiguration.Namespace))
			Expect(schedule.Spec.Pipeline.Name).To(Equal(runConfiguration.Spec.Run.Pipeline.Name))
			Expect(schedule.Spec.Pipeline.Version).To(Equal(runConfiguration.Status.ObservedPipelineVersion))
			Expect(schedule.Spec.RuntimeParameters).To(Equal(resolvedParameters))
			Expect(schedule.Spec.ExperimentName).To(Equal(runConfiguration.Spec.Run.ExperimentName))
			Expect(schedule.Spec.Schedule).To(Equal(runConfiguration.Spec.Triggers.Schedules[i]))
			Expect(metav1.IsControlledBy(&schedule, runConfiguration)).To(BeTrue())
			Expect(schedule.Spec.Provider).To(Equal(runConfiguration.Spec.Run.Provider))
		}
	})
})

var _ = Context("updateRcTriggers", PropertyBased, func() {
	It("sets the pipeline trigger status", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(apis.RandomLowercaseString())
		runConfiguration.Spec.Triggers.OnChange = []pipelineshub.OnChangeType{
			pipelineshub.OnChangeTypes.Pipeline,
		}
		runConfiguration.Status.ObservedPipelineVersion = apis.RandomString()
		rcr := RunConfigurationReconciler{}
		Expect(rcr.updateRcTriggers(*runConfiguration).TriggeredPipelineVersion).To(Equal(runConfiguration.Status.ObservedPipelineVersion))
	})

	It("sets the runSpec trigger status", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(apis.RandomLowercaseString())
		runConfiguration.Spec.Triggers.OnChange = []pipelineshub.OnChangeType{
			pipelineshub.OnChangeTypes.RunSpec,
		}
		rcr := RunConfigurationReconciler{}
		Expect(rcr.updateRcTriggers(*runConfiguration).Triggers.RunSpec.Version).To(Equal(runConfiguration.Spec.Run.ComputeVersion()))
	})

	It("sets the runConfigurations trigger status", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(apis.RandomLowercaseString())
		runConfiguration.Spec.Triggers.RunConfigurations = apis.RandomList(apis.RandomString)
		runConfiguration.Status.Dependencies.RunConfigurations = make(map[string]pipelineshub.RunReference)
		for _, rc := range runConfiguration.Spec.Triggers.RunConfigurations {
			runConfiguration.Status.Dependencies.RunConfigurations[rc] = pipelineshub.RunReference{
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
		runConfiguration := pipelineshub.RandomRunConfiguration(apis.RandomLowercaseString())
		rcr := RunConfigurationReconciler{}
		updatedStatus := rcr.updateRcTriggers(*runConfiguration)
		updatedStatus.Triggers = runConfiguration.Status.Triggers
		updatedStatus.TriggeredPipelineVersion = runConfiguration.Status.TriggeredPipelineVersion
		Expect(updatedStatus).To(Equal(runConfiguration.Status))
	})
})
