package v1alpha3

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha5"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Experiment) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Experiment)

	v1alpha4remainder := v1alpha4.ResourceConversionRemainder{}
	if err := pipelines.RetrieveAndUnsetConversionAnnotations(src, &v1alpha4remainder); err != nil {
		return err
	}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Description = src.Spec.Description
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.ProviderId = hub.ProviderAndId{
		Provider: v1alpha4remainder.Provider,
		Id:       src.Status.KfpId,
	}
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version

	return nil
}

func (dst *Experiment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Experiment)

	v1alpha4remainder := v1alpha4.ResourceConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Description = src.Spec.Description
	dst.Status.SynchronizationState = src.Status.SynchronizationState
	dst.Status.KfpId = src.Status.ProviderId.Id
	dst.Status.ObservedGeneration = src.Status.ObservedGeneration
	dst.Status.Version = src.Status.Version

	v1alpha4remainder.Provider = src.Status.ProviderId.Provider
	if err := pipelines.SetConversionAnnotations(dst, &v1alpha4remainder); err != nil {
		return err
	}

	return nil
}
