package v1alpha4

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = hub.ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version
	dst.Spec.Env = src.Spec.Env
	dst.Spec.BeamArgs = src.Spec.BeamArgs
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = ProviderAndId{
		Provider: src.Status.ProviderId.Provider,
		Id:       src.Status.ProviderId.Id,
	}
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version
	dst.Spec.Env = src.Spec.Env
	dst.Spec.BeamArgs = src.Spec.BeamArgs
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents

	return nil
}
