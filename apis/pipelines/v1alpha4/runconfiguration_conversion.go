package v1alpha4

import (
	"fmt"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Run.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	if src.Spec.Schedule != "" {
		dst.Spec.Triggers = []hub.Trigger{{Schedule: &hub.ScheduleTrigger{CronExpression: src.Spec.Schedule}}}
	}
	dst.Spec.Run.ExperimentName = src.Spec.ExperimentName
	dst.Status = hub.RunConfigurationStatus{
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
		SynchronizationState:    src.Status.SynchronizationState,
		Provider:                src.Status.ProviderId.Provider,
		ObservedGeneration:      src.Status.ObservedGeneration,
	}

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)

	var scheduleTrigger *hub.ScheduleTrigger
	switch len(src.Spec.Triggers) {
	case 0:
		scheduleTrigger = &hub.ScheduleTrigger{}
	case 1:
		scheduleTrigger = src.Spec.Triggers[0].Schedule
		if scheduleTrigger == nil {
			return fmt.Errorf("conversion only supported for schedule triggers")
		}
	default:
		return fmt.Errorf("conversion only supported for at most one trigger")
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = src.Spec.Run.RuntimeParameters
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Run.Pipeline.Name, Version: src.Spec.Run.Pipeline.Version}
	dst.Spec.Schedule = scheduleTrigger.CronExpression
	dst.Spec.ExperimentName = src.Spec.Run.ExperimentName

	dst.Status = RunConfigurationStatus{
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
		Status: Status{
			SynchronizationState: src.Status.SynchronizationState,
			ProviderId: ProviderAndId{
				Provider: src.Status.Provider,
			},
			ObservedGeneration: src.Status.ObservedGeneration,
		},
	}

	return nil
}
