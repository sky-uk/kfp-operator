package v1alpha4

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	v1alpha5remainder := hub.ResourceConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha5remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Run.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	if src.Spec.Schedule != "" {
		dst.Spec.Triggers = []hub.Trigger{{Schedule: &hub.ScheduleTrigger{CronExpression: src.Spec.Schedule}}}
	}
	dst.Spec.Triggers = append(dst.Spec.Triggers, v1alpha5remainder.Triggers...)
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

	v1alpha5remainder := hub.ResourceConversionRemainder{}

	var scheduleTrigger *hub.ScheduleTrigger

	for _, trigger := range src.Spec.Triggers {
		if trigger.Schedule != nil && scheduleTrigger == nil {
			scheduleTrigger = trigger.Schedule
		} else {
			v1alpha5remainder.Triggers = append(v1alpha5remainder.Triggers, trigger)
		}
	}

	if scheduleTrigger == nil {
		scheduleTrigger = &hub.ScheduleTrigger{}
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

	return pipelines.SetConversionAnnotations(dst, &v1alpha5remainder)
}
