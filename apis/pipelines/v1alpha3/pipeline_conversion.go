package v1alpha3

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)

	v1alpha4remainder := v1alpha4.ResourceConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha4remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = hub.ProviderAndId{
		Provider: v1alpha4remainder.Provider,
		Id:       src.Status.KfpId,
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

	v1alpha4remainder := v1alpha4.ResourceConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.KfpId = src.Status.ProviderId.Id
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version
	dst.Spec.Env = src.Spec.Env
	dst.Spec.BeamArgs = src.Spec.BeamArgs
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents

	v1alpha4remainder.Provider = src.Status.ProviderId.Provider
	if err := pipelines.SetConversionAnnotations(dst, &v1alpha4remainder); err != nil {
		return err
	}

	return nil
}
