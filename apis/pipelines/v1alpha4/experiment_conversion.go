package v1alpha4

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Experiment) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Experiment)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Description = src.Spec.Description
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = hub.ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version

	return nil
}

func (dst *Experiment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Experiment)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Description = src.Spec.Description
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version

	return nil
}
