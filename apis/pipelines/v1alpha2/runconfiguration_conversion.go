package v1alpha2

import (
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
	v1alpha5remainder := hub.RunConfigurationConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha3remainder, &v1alpha4remainder, &v1alpha5remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.RuntimeParameters = hub.MergeRuntimeParameters(append(v1alpha4.MapToNamedValues(src.Spec.RuntimeParameters), v1alpha3remainder.RuntimeParameters...), v1alpha5remainder.ValueFromParameters)
	dst.Spec.Run.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Triggers = v1alpha5remainder.Triggers
	if src.Spec.Schedule != "" {
		dst.Spec.Triggers.Schedules = append([]string{src.Spec.Schedule}, dst.Spec.Triggers.Schedules...)
	}
	dst.Spec.Run.Artifacts = v1alpha5remainder.Artifacts
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

	v1alpha3remainder := v1alpha3.RunConfigurationConversionRemainder{}
	v1alpha4remainder := v1alpha4.ResourceConversionRemainder{}
	v1alpha5remainder := hub.RunConfigurationConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	valueParameters, valueFromParameters := hub.SplitRunTimeParameters(src.Spec.Run.RuntimeParameters)
	dst.Spec.RuntimeParameters, v1alpha3remainder.RuntimeParameters = v1alpha4.NamedValuesToMap(valueParameters)
	v1alpha5remainder.ValueFromParameters = valueFromParameters
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Run.Pipeline.Name, Version: src.Spec.Run.Pipeline.Version}
	v1alpha5remainder.Triggers = src.Spec.Triggers
	if len(src.Spec.Triggers.Schedules) > 0 {
		dst.Spec.Schedule = v1alpha5remainder.Triggers.Schedules[0]
		v1alpha5remainder.Triggers.Schedules = v1alpha5remainder.Triggers.Schedules[1:]
	}
	v1alpha5remainder.Artifacts = src.Spec.Run.Artifacts
	dst.Spec.ExperimentName = src.Spec.Run.ExperimentName
	dst.Status = RunConfigurationStatus{
		Status: Status{
			SynchronizationState: src.Status.SynchronizationState,
			ObservedGeneration:   src.Status.ObservedGeneration,
		},
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
	}

	v1alpha4remainder.Provider = src.Status.Provider
	if err := pipelines.SetConversionAnnotations(dst, &v1alpha3remainder, &v1alpha4remainder, &v1alpha5remainder); err != nil {
		return err
	}

	return nil
}
