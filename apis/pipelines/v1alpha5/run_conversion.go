package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Run) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Run)

	err := pipelines.TransformInto(src, &dst)

	dst.Status.Status = hub.Status{
		ProviderId: hub.ProviderAndId{
			Provider: src.Status.ProviderId.Provider,
			Id:       src.Status.ProviderId.Id,
		},
		SynchronizationState: src.Status.SynchronizationState,
		Version:              src.Status.Version,
		ObservedGeneration:   src.Status.ObservedGeneration,
		Conditions:           hub.Conditions(src.Status.Conditions),
	}

	if err != nil {
		return err
	}

	dst.Spec.Provider = getProviderAnnotation(src)
	removeProviderAnnotation(dst)

	return nil
}

func (dst *Run) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Run)

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}
	setProviderAnnotation(src.Spec.Provider, &dst.ObjectMeta)

	dst.Status.ProviderId = ProviderAndId{
		Provider: src.Status.Status.ProviderId.Provider,
		Id:       src.Status.Status.ProviderId.Id,
	}
	dst.Status.SynchronizationState = src.Status.Status.SynchronizationState
	dst.Status.Version = src.Status.Status.Version
	dst.Status.ObservedGeneration = src.Status.Status.ObservedGeneration
	dst.Status.Conditions = Conditions(src.Status.Status.Conditions)

	return nil
}
