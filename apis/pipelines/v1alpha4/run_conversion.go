package v1alpha4

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Run) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Run)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = hub.PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = hub.ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.ObservedPipelineVersion = src.Status.ObservedPipelineVersion
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version

	return nil
}

func (dst *Run) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Run)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Pipeline = PipelineIdentifier{
		Name:    src.Spec.Pipeline.Name,
		Version: src.Spec.Pipeline.Version,
	}
	dst.Spec.RuntimeParameters = src.Spec.RuntimeParameters
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.ObservedPipelineVersion = src.Status.ObservedPipelineVersion
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version

	return nil
}
