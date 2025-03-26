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
		remainder.ProviderStatusNamespace,
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

	dst.Spec.Run.Provider = src.Spec.Run.Provider.Name
	dst.Status.Provider = src.Status.Provider.Name
	remainder.ProviderNamespace = src.Spec.Run.Provider.Namespace
	remainder.ProviderStatusNamespace = src.Status.Provider.Namespace

	dst.TypeMeta.APIVersion = dstApiVersion

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
