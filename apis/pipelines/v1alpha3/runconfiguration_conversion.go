package v1alpha3

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	v1alpha4remainder := v1alpha4.ResourceConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha4remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Run.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	if src.Spec.Schedule != "" {
		dst.Spec.Triggers = []hub.Trigger{{Schedule: &hub.ScheduleTrigger{
			CronExpression: src.Spec.Schedule,
		}}}
	}
	dst.Spec.Run.ExperimentName = src.Spec.ExperimentName
	dst.Status = hub.RunConfigurationStatus{
		SynchronizationState:    src.Status.SynchronizationState,
		Provider:                v1alpha4remainder.Provider,
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
		ObservedGeneration:      src.Status.ObservedGeneration,
	}

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)

	v1alpha4remainder := v1alpha4.ResourceConversionRemainder{}

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
		Status: Status{
			SynchronizationState: src.Status.SynchronizationState,
			ObservedGeneration:   src.Status.ObservedGeneration,
		},
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
	}

	v1alpha4remainder.Provider = src.Status.Provider
	if err := pipelines.SetConversionAnnotations(dst, &v1alpha4remainder); err != nil {
		return err
	}

	return nil
}
