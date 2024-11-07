package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec = hub.PipelineSpec{
		Provider:      getProviderAnnotation(src),
		Image:         src.Spec.Image,
		TfxComponents: src.Spec.TfxComponents,
		Env:           src.Spec.Env,
		BeamArgs:      src.Spec.BeamArgs,
	}
	removeProviderAnnotation(dst)
	dst.Status = hub.Status{
		ProviderId: hub.ProviderAndId{
			Provider: src.Status.ProviderId.Provider,
			Id:       src.Status.ProviderId.Id,
		},
		SynchronizationState: src.Status.SynchronizationState,
		Version:              src.Status.Version,
		ObservedGeneration:   src.Status.ObservedGeneration,
		Conditions:           hub.Conditions(src.Status.Conditions),
	}
	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)

	dst.ObjectMeta = src.ObjectMeta
	setProviderAnnotation(src.Spec.Provider, &dst.ObjectMeta)

	dst.Spec = PipelineSpec{
		Image:         src.Spec.Image,
		TfxComponents: src.Spec.TfxComponents,
		Env:           src.Spec.Env,
		BeamArgs:      src.Spec.BeamArgs,
	}
	dst.Status = Status{
		ProviderId: ProviderAndId{
			Provider: src.Status.ProviderId.Provider,
			Id:       src.Status.ProviderId.Id,
		},
		SynchronizationState: src.Status.SynchronizationState,
		Version:              src.Status.Version,
		ObservedGeneration:   src.Status.ObservedGeneration,
		Conditions:           Conditions(src.Status.Conditions),
	}
	return nil
}
