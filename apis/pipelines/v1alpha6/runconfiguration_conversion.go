package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunConfiguration)
	dstApiVersion := dst.APIVersion
	remainder := RunConfigurationConversionRemainder{}

	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}
	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.Run.Provider = convertProviderTo(
		src.Spec.Run.Provider,
		remainder.ProviderNamespace,
	)
	dst.Status.Provider = convertProviderTo(
		src.Status.Provider,
		remainder.ProviderNamespace,
	)
	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)
	dstApiVersion := dst.APIVersion
	remainder := RunConfigurationConversionRemainder{}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.Run.Provider = convertProviderFrom4(src.Spec.Run.Provider, &remainder)
	dst.Status.Provider = convertProviderFrom4(src.Status.Provider, &remainder)
	dst.TypeMeta.APIVersion = dstApiVersion

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
