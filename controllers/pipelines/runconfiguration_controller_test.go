//go:build unit

package pipelines

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/samber/lo"
	"github.com/sky-uk/kfp-operator/apis"
	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/controllers/pipelines/internal/workflowfactory"
	"github.com/sky-uk/kfp-operator/pkg/common"
	"github.com/sky-uk/kfp-operator/pkg/common/triggers"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Context("aggregateState", func() {
	const updatingMessage = "Waiting for all dependant runschedules to be in a state of Succeeded"
	DescribeTable("calculates based on sub states", func(subStates []apis.SynchronizationState, expectedState apis.SynchronizationState, expectedMessage string) {
		runSchedules := make([]pipelineshub.RunSchedule, len(subStates))
		for i, state := range subStates {
			runSchedules[i] = pipelineshub.RunSchedule{
				Status: pipelineshub.Status{
					Conditions: []metav1.Condition{
						{
							Type:    apis.ConditionTypes.SynchronizationSucceeded,
							Message: string(state),
							Reason:  string(state),
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
		Entry(nil, []apis.SynchronizationState{apis.Updating, apis.Failed}, apis.Updating, updatingMessage),
		Entry(nil, []apis.SynchronizationState{apis.Deleting, apis.Failed}, apis.Updating, updatingMessage),
		Entry(nil, []apis.SynchronizationState{apis.Deleted, apis.Failed}, apis.Updating, updatingMessage),
		Entry(nil, []apis.SynchronizationState{"", apis.Failed}, apis.Updating, updatingMessage),
		Entry(nil, []apis.SynchronizationState{apis.Succeeded}, apis.Succeeded, ""),
	)
})

var _ = Context("constructRunForRunConfiguration", PropertyBased, func() {
	Expect(pipelineshub.AddToScheme(scheme.Scheme)).To(Succeed())
	rcr := RunConfigurationReconciler{
		Scheme: scheme.Scheme,
	}

	It("propagates the runconfiguration's name", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
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
		runConfiguration := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
		runConfiguration.Spec.Triggers = pipelineshub.Triggers{Schedules: apis.RandomList(pipelineshub.RandomSchedule)}
		resolvedParameters := lo.Map(runConfiguration.Spec.Run.Parameters, func(p pipelineshub.Parameter, _ int) apis.NamedValue {
			return apis.NamedValue{Name: p.Name, Value: p.Value}
		})

		runSchedules, err := rcr.constructRunSchedulesForTriggers(runConfiguration, resolvedParameters)
		Expect(err).NotTo(HaveOccurred())

		for i, schedule := range runSchedules {
			Expect(schedule.Namespace).To(Equal(runConfiguration.Namespace))
			Expect(schedule.Spec.Pipeline.Name).To(Equal(runConfiguration.Spec.Run.Pipeline.Name))
			Expect(schedule.Spec.Pipeline.Version).To(Equal(runConfiguration.Status.Dependencies.Pipeline.Version))
			Expect(schedule.Spec.Parameters).To(Equal(resolvedParameters))
			Expect(schedule.Spec.ExperimentName).To(Equal(runConfiguration.Spec.Run.ExperimentName))
			Expect(schedule.Spec.Schedule).To(Equal(runConfiguration.Spec.Triggers.Schedules[i]))
			Expect(metav1.IsControlledBy(&schedule, runConfiguration)).To(BeTrue())
			Expect(schedule.Spec.Provider).To(Equal(runConfiguration.Spec.Run.Provider))
		}
	})
})

var _ = Context("updateRcTriggers", PropertyBased, func() {
	It("sets the pipeline trigger status", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
		runConfiguration.Spec.Triggers.OnChange = []pipelineshub.OnChangeType{
			pipelineshub.OnChangeTypes.Pipeline,
		}
		runConfiguration.Status.Dependencies.Pipeline.Version = apis.RandomString()
		rcr := RunConfigurationReconciler{}
		Expect(rcr.updateRcTriggers(*runConfiguration).Triggers.Pipeline.Version).To(Equal(runConfiguration.Status.Dependencies.Pipeline.Version))
	})

	It("sets the runSpec trigger status", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
		runConfiguration.Spec.Triggers.OnChange = []pipelineshub.OnChangeType{
			pipelineshub.OnChangeTypes.RunSpec,
		}
		rcr := RunConfigurationReconciler{}
		Expect(rcr.updateRcTriggers(*runConfiguration).Triggers.RunSpec.Version).To(Equal(runConfiguration.Spec.Run.ComputeVersion()))
	})

	It("sets the runConfigurations trigger status", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
		runConfiguration.Spec.Triggers.RunConfigurations = apis.RandomList(common.RandomNamespacedName)
		runConfiguration.Status.Dependencies.RunConfigurations = make(map[string]pipelineshub.RunReference)
		for _, rc := range runConfiguration.Spec.Triggers.RunConfigurations {
			rcNamespacedName, err := rc.String()
			Expect(err).NotTo(HaveOccurred())
			runConfiguration.Status.Dependencies.RunConfigurations[rcNamespacedName] = pipelineshub.RunReference{
				ProviderId: apis.RandomString(),
			}
		}
		rcr := RunConfigurationReconciler{}
		updatedStatus := rcr.updateRcTriggers(*runConfiguration)
		for _, rc := range runConfiguration.Spec.Triggers.RunConfigurations {
			rcNamespacedName, err := rc.String()
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedStatus.Triggers.RunConfigurations[rcNamespacedName].ProviderId).To(Equal(runConfiguration.Status.Dependencies.RunConfigurations[rcNamespacedName].ProviderId))
		}
		Expect(updatedStatus.Triggers.RunConfigurations).To(HaveLen(len(runConfiguration.Spec.Triggers.RunConfigurations)))
	})

	It("retains other fields", func() {
		runConfiguration := pipelineshub.RandomRunConfiguration(common.RandomNamespacedName())
		rcr := RunConfigurationReconciler{}
		updatedStatus := rcr.updateRcTriggers(*runConfiguration)
		updatedStatus.Triggers = runConfiguration.Status.Triggers
		updatedStatus.Triggers.Pipeline.Version = runConfiguration.Status.Triggers.Pipeline.Version
		Expect(updatedStatus).To(Equal(runConfiguration.Status))
	})
})

var _ = Context("IdentifyRunTriggerReason", func() {
	reconciler := &RunConfigurationReconciler{}

	rcName := "run-config"
	rcNamespace := "namespace"
	pipelineName := "pipeline"

	rc := &pipelineshub.RunConfiguration{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rcName,
			Namespace: rcNamespace,
		},
		Spec: pipelineshub.RunConfigurationSpec{
			Run: pipelineshub.RunSpec{
				Pipeline: pipelineshub.PipelineIdentifier{
					Name:    pipelineName,
					Version: "0.0.1",
				},
			},
		},
	}

	DescribeTable("determines correct trigger reason",
		func(status, old pipelineshub.RunConfigurationStatus, expected *triggers.Indicator) {
			rc.Status = status
			indicator := reconciler.IdentifyRunTriggerReason(rc, old)
			Expect(indicator).To(Equal(expected))
		},

		Entry("returns OnChangeRunSpec when RunSpec version changes",
			pipelineshub.RunConfigurationStatus{
				Triggers: pipelineshub.TriggersStatus{
					RunSpec: pipelineshub.RunSpecTriggerStatus{Version: "0.0.1"},
				},
			},
			pipelineshub.RunConfigurationStatus{
				Triggers: pipelineshub.TriggersStatus{
					RunSpec: pipelineshub.RunSpecTriggerStatus{Version: "0.0.2"},
				},
			},
			&triggers.Indicator{
				Type:            triggers.OnChangeRunSpec,
				Source:          rcName,
				SourceNamespace: rcNamespace,
			},
		),

		Entry("returns OnChangePipeline when Pipeline version changes",
			pipelineshub.RunConfigurationStatus{
				Triggers: pipelineshub.TriggersStatus{
					Pipeline: pipelineshub.PipelineTriggerStatus{Version: "0.0.1"},
				},
			},
			pipelineshub.RunConfigurationStatus{
				Triggers: pipelineshub.TriggersStatus{
					Pipeline: pipelineshub.PipelineTriggerStatus{Version: "0.0.2"},
				},
			},
			&triggers.Indicator{
				Type:            triggers.OnChangePipeline,
				Source:          pipelineName,
				SourceNamespace: rcNamespace,
			},
		),

		Entry("returns RunConfiguration when RunConfigurations differ",
			pipelineshub.RunConfigurationStatus{
				Triggers: pipelineshub.TriggersStatus{
					RunConfigurations: map[string]pipelineshub.TriggeredRunReference{
						"namespace2/run-configuration-dependent": {ProviderId: "0.0.1"},
					},
				},
			},
			pipelineshub.RunConfigurationStatus{
				Triggers: pipelineshub.TriggersStatus{
					RunConfigurations: map[string]pipelineshub.TriggeredRunReference{
						"namespace2/run-configuration-dependent": {ProviderId: "0.0.2"},
					},
				},
			},
			&triggers.Indicator{
				Type:            triggers.RunConfiguration,
				Source:          "run-configuration-dependent",
				SourceNamespace: "namespace2",
			},
		),

		Entry("returns nil when all versions and configs are the same",
			pipelineshub.RunConfigurationStatus{
				Triggers: pipelineshub.TriggersStatus{
					RunSpec:  pipelineshub.RunSpecTriggerStatus{Version: "0.0.1"},
					Pipeline: pipelineshub.PipelineTriggerStatus{Version: "0.0.1"},
					RunConfigurations: map[string]pipelineshub.TriggeredRunReference{
						"namespace2/run-configuration-dependent": {ProviderId: "0.0.1"},
					},
				},
			},
			pipelineshub.RunConfigurationStatus{
				Triggers: pipelineshub.TriggersStatus{
					RunSpec:  pipelineshub.RunSpecTriggerStatus{Version: "0.0.1"},
					Pipeline: pipelineshub.PipelineTriggerStatus{Version: "0.0.1"},
					RunConfigurations: map[string]pipelineshub.TriggeredRunReference{
						"namespace2/run-configuration-dependent": {ProviderId: "0.0.1"},
					},
				},
			},
			nil,
		),
	)
})
