package v1alpha4

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunSchedule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunSchedule)

	v1alpha5remainder := hub.RunConfigurationConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha5remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = hub.MergeRuntimeParameters(src.Spec.RuntimeParameters, v1alpha5remainder.ValueFromParameters)
	dst.Spec.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
	dst.Spec.Artifacts = v1alpha5remainder.Artifacts
	dst.Spec.ExperimentName = src.Spec.ExperimentName

	dst.Status.Version = src.Status.Version
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.ProviderId.Provider = src.Status.ProviderId.Provider
	dst.Status.ProviderId.Id = src.Status.ProviderId.Id

	return nil
}

func (dst *RunSchedule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunSchedule)

	v1alpha5remainder := hub.RunConfigurationConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters, v1alpha5remainder.ValueFromParameters = hub.SplitRunTimeParameters(src.Spec.RuntimeParameters)
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	v1alpha5remainder.Artifacts = src.Spec.Artifacts
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.Schedule = src.Spec.Schedule

	dst.Status.Version = src.Status.Version
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.ProviderId.Provider = src.Status.ProviderId.Provider
	dst.Status.ProviderId.Id = src.Status.ProviderId.Id

	return pipelines.SetConversionAnnotations(dst, &v1alpha5remainder)
}
