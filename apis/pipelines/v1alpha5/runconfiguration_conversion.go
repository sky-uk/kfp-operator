package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)
	v1alpha6Remainder := hub.RunConfigurationConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha6Remainder); err != nil {
		return err
	}
	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Run.Pipeline.Name,
		Version: src.Spec.Run.Pipeline.Version,
	}
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName
	dst.Spec.Run.RuntimeParameters = convertRuntimeParametersTo(
		src.Spec.Run.RuntimeParameters,
	)
	dst.Spec.Run.Artifacts = convertArtifactsTo(src.Spec.Run.Artifacts)
	dst.Spec.Triggers = convertTriggersTo(
		src.Spec.Triggers,
		v1alpha6Remainder.Schedules,
	)
	dst.Status = hub.RunConfigurationStatus{
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
		SynchronizationState:    src.Status.SynchronizationState,
		Provider:                src.Status.Provider,
		ObservedGeneration:      src.Status.ObservedGeneration,
	}

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)
	v1alpha6Remainder := hub.RunConfigurationConversionRemainder{}
	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Run.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Run.Pipeline.Name,
		Version: src.Spec.Run.Pipeline.Version,
	}
	dst.Spec.Run.ExperimentName = src.Spec.Run.ExperimentName
	// TODO: missing ValueFrom. nil ptr dereference
	dst.Spec.Run.RuntimeParameters = convertRuntimeParametersFrom(src.Spec.Run.RuntimeParameters)
	dst.Spec.Run.Artifacts = convertArtifactsFrom(src.Spec.Run.Artifacts)
	dst.Spec.Triggers = convertTriggersFrom(src.Spec.Triggers)
	// TODO: convertTriggersFrom should also mutate the remainder
	v1alpha6Remainder.Schedules = src.Spec.Triggers.Schedules
	dst.Status = RunConfigurationStatus{
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
		SynchronizationState:    src.Status.SynchronizationState,
		Provider:                src.Status.Provider,
		ObservedGeneration:      src.Status.ObservedGeneration,
	}
	return pipelines.SetConversionAnnotations(dst, &v1alpha6Remainder)
}
