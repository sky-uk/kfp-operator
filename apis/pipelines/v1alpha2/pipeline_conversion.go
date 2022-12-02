package v1alpha2

import (
	pipelinesv1alpha3 "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)

	v1alpha3remainder := pipelinesv1alpha3.PipelineConversionRemainder{}
	v1alpha4remainder := hub.ResourceConversionRemainder{}
	if err := hub.RetrieveAndUnsetConversionAnnotations(src, &v1alpha3remainder, &v1alpha4remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = hub.ProviderId{
		Provider: v1alpha4remainder.Provider,
		Id:       src.Status.KfpId,
	}
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version
	dst.Spec.Env = hub.MapToNamedValues(src.Spec.Env)
	dst.Spec.BeamArgs = hub.MapToNamedValues(src.Spec.BeamArgs)
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents
	dst.Spec.BeamArgs = append(dst.Spec.BeamArgs, v1alpha3remainder.BeamArgs...)
	dst.Spec.Env = append(dst.Spec.Env, v1alpha3remainder.Env...)

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)

	v1alpha3remainder := pipelinesv1alpha3.PipelineConversionRemainder{}
	v1alpha4remainder := hub.ResourceConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.KfpId = src.Status.ProviderId.Id
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version
	dst.Spec.Env, v1alpha3remainder.Env = hub.NamedValuesToMap(src.Spec.Env)
	dst.Spec.BeamArgs, v1alpha3remainder.BeamArgs = hub.NamedValuesToMap(src.Spec.BeamArgs)
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents

	v1alpha4remainder.Provider = src.Status.ProviderId.Provider
	if err := hub.SetConversionAnnotations(dst, &v1alpha3remainder, &v1alpha4remainder); err != nil {
		return err
	}

	return nil
}
