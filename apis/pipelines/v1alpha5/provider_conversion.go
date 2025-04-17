package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Provider) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Provider)
	dstApiVersion := dst.APIVersion

	remainder := ProviderConversionRemainder{}

	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}
	dst.Spec.ServiceImage = remainder.ServiceImage
	remainder.ServiceImage = ""
	remainder.Image = src.Spec.Image

	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Spec.Parameters = src.Spec.Parameters
	return pipelines.SetConversionAnnotations(dst, &remainder)
}

func (dst *Provider) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Provider)
	dstApiVersion := dst.APIVersion

	remainder := ProviderConversionRemainder{}
	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	status := src.Status.Conditions.GetSyncStateFromReason()

	dst.Status.SynchronizationState = status
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Spec.Image = remainder.Image
	dst.Spec.Parameters = src.Spec.Parameters

	remainder.Image = ""
	remainder.ServiceImage = src.Spec.ServiceImage

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
