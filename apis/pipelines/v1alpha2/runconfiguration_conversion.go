package v1alpha2

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	v1alpha3remainder := v1alpha3.RunConfigurationConversionRemainder{}
	v1alpha4remainder := v1alpha4.ResourceConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha3remainder, &v1alpha4remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = append(v1alpha4.MapToNamedValues(src.Spec.RuntimeParameters), v1alpha3remainder.RuntimeParameters...)
	dst.Spec.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	if src.Spec.Schedule != "" {
		dst.Spec.Triggers = []hub.Trigger{{Type: hub.TriggerTypes.Schedule, CronExpression: src.Spec.Schedule}}
	}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
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

	v1alpha3remainder := v1alpha3.RunConfigurationConversionRemainder{}
	v1alpha4remainder := v1alpha4.ResourceConversionRemainder{}

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
	dst.Spec.RuntimeParameters, v1alpha3remainder.RuntimeParameters = v1alpha4.NamedValuesToMap(src.Spec.RuntimeParameters)
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = trigger.CronExpression
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status = RunConfigurationStatus{
		Status: Status{
			SynchronizationState: src.Status.SynchronizationState,
			ObservedGeneration:   src.Status.ObservedGeneration,
		},
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
	}

	v1alpha4remainder.Provider = src.Status.Provider
	if err := pipelines.SetConversionAnnotations(dst, &v1alpha3remainder, &v1alpha4remainder); err != nil {
		return err
	}

	return nil
}
