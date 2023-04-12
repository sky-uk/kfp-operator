package v1alpha3

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)

	v1alpha4remainder := hub.ResourceConversionRemainder{}
	if err := hub.RetrieveAndUnsetConversionAnnotations(src, &v1alpha4remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Pipeline = hub.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status = hub.RunConfigurationStatus{
		Status: hub.Status{
			ProviderId: hub.ProviderAndId{
				Provider: v1alpha4remainder.Provider,
				Id:       src.Status.KfpId,
			},
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

	v1alpha4remainder := hub.ResourceConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status = RunConfigurationStatus{
		Status: Status{
			KfpId:                src.Status.ProviderId.Id,
			SynchronizationState: src.Status.SynchronizationState,
			Version:              src.Status.Version,
			ObservedGeneration:   src.Status.ObservedGeneration,
		},
		ObservedPipelineVersion: src.Status.ObservedPipelineVersion,
	}

	v1alpha4remainder.Provider = src.Status.ProviderId.Provider
	if err := hub.SetConversionAnnotations(dst, &v1alpha4remainder); err != nil {
		return err
	}

	return nil
}
