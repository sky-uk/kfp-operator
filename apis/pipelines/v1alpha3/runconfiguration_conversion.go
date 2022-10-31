package v1alpha3

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status = hub.RunConfigurationStatus{
		Status: hub.Status{
			ProviderId:           src.Status.KfpId,
			SynchronizationState: src.Status.SynchronizationState,
			Version:              src.Status.Version,
			ObservedGeneration:   src.Status.ObservedGeneration,
		},
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
	}

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status = RunConfigurationStatus{
		Status: Status{
			KfpId:                src.Status.ProviderId,
			SynchronizationState: src.Status.SynchronizationState,
			Version:              src.Status.Version,
			ObservedGeneration:   src.Status.ObservedGeneration,
		},
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
	}

	return nil
}
