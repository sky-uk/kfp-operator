package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Run) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Run)
	dstApiVersion := dst.APIVersion
	remainder := RunConversionRemainder{}

	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}
	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.Provider = convertProviderTo(
		getProviderAnnotation(src),
		remainder.ProviderNamespace,
	)
	dst.Status.Provider = convertProviderAndIdTo(
		src.Status.ProviderId,
		remainder.ProviderStatusNamespace,
	)

	removeProviderAnnotation(dst)

	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}

func (dst *Run) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Run)
	dstApiVersion := dst.APIVersion
	remainder := RunConversionRemainder{}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	setProviderAnnotation(src.Spec.Provider.Name, &dst.ObjectMeta)
	remainder.ProviderNamespace = src.Spec.Provider.Namespace
	remainder.ProviderStatusNamespace = src.Status.Provider.Name.Namespace
	dst.Status.SynchronizationState = src.Status.Conditions.GetSyncStateFromReason()
	dst.Status.ProviderId = convertProviderAndIdFrom(src.Status.Provider)
	dst.TypeMeta.APIVersion = dstApiVersion

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
