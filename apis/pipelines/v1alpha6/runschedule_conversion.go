package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunSchedule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunSchedule)
	dstApiVersion := dst.APIVersion
	remainder := RunScheduleConversionRemainder{}

	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}
	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	dst.Spec.Provider = convertProviderTo(
		src.Spec.Provider,
		remainder.ProviderNamespace,
	)
	dst.Status.Provider.Name = convertProviderTo(
		src.Status.Provider.Name,
		remainder.ProviderStatusNamespace,
	)
	dst.TypeMeta.APIVersion = dstApiVersion

	dst.Spec.Parameters = src.Spec.RuntimeParameters

	return nil
}

func (dst *RunSchedule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunSchedule)
	dstApiVersion := dst.APIVersion
	remainder := RunScheduleConversionRemainder{}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.Provider = src.Spec.Provider.Name
	dst.Spec.RuntimeParameters = src.Spec.Parameters
	dst.Status.Provider.Name = src.Status.Provider.Name.Name
	remainder.ProviderNamespace = src.Spec.Provider.Namespace
	remainder.ProviderStatusNamespace = src.Status.Provider.Name.Namespace
	dst.Status.SynchronizationState = src.Status.Conditions.GetSyncStateFromReason()
	dst.TypeMeta.APIVersion = dstApiVersion

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
