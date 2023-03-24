package v1alpha4

import (
	"fmt"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	if src.Spec.Schedule != "" {
		dst.Spec.Triggers = []hub.Trigger{{Type: hub.TriggerTypes.Schedule, CronExpression: src.Spec.Schedule}}
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
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

	var trigger hub.Trigger
	switch len(src.Spec.Triggers) {
	case 0:
		trigger = hub.Trigger{}
	case 1:
		trigger = src.Spec.Triggers[0]
		if trigger.Type != hub.TriggerTypes.Schedule {
			return fmt.Errorf("conversion only supported for schedule triggers")
		}
	default:
		return fmt.Errorf("conversion only supported for at most one trigger")
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = trigger.CronExpression
	dst.Spec.ExperimentName = src.Spec.ExperimentName

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
