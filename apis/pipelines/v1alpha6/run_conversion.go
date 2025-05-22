package v1alpha6

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

	dst.Status.Dependencies.Pipeline.Version = src.Status.ObservedPipelineVersion

	dst.Spec.Provider = convertProviderTo(
		src.Spec.Provider,
		remainder.ProviderNamespace,
	)
	dst.Status.Provider.Name = convertProviderTo(
		src.Status.Provider.Name,
		remainder.ProviderStatusNamespace,
	)
	dst.TypeMeta.APIVersion = dstApiVersion

	if len(dst.Spec.RuntimeParameters) > 0 {
		dst.Spec.Parameters = dst.Spec.RuntimeParameters
		dst.Spec.RuntimeParameters = nil
	}

	return nil
}

func (dst *Run) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Run)
	dstApiVersion := dst.APIVersion
	remainder := RunConversionRemainder{}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Status.ObservedPipelineVersion = src.Status.Dependencies.Pipeline.Version
	dst.Spec.Provider = src.Spec.Provider.Name
	dst.Status.Provider.Name = src.Status.Provider.Name.Name
	remainder.ProviderNamespace = src.Spec.Provider.Namespace
	remainder.ProviderStatusNamespace = src.Status.Provider.Name.Namespace
	dst.Status.SynchronizationState = src.Status.Conditions.GetSyncStateFromReason()
	dst.TypeMeta.APIVersion = dstApiVersion

	if len(dst.Spec.Parameters) > 0 {
		dst.Spec.RuntimeParameters = dst.Spec.Parameters
		dst.Spec.Parameters = nil
	}

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
