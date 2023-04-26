package v1alpha4

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunSchedule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunSchedule)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
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

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Spec.Schedule = src.Spec.Schedule

	dst.Status.Version = src.Status.Version
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.ProviderId.Provider = src.Status.ProviderId.Provider
	dst.Status.ProviderId.Id = src.Status.ProviderId.Id

	return nil
}
