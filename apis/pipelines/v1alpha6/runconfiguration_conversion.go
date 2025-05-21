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

	dst.Status.Dependencies.Pipeline.Version = src.Status.ObservedPipelineVersion
	dst.Status.Triggers.Pipeline.Version = src.Status.TriggeredPipelineVersion

	dst.Spec.Run.Provider = convertProviderTo(
		src.Spec.Run.Provider,
		remainder.ProviderNamespace,
	)
	dst.Status.Provider = convertProviderTo(
		src.Status.Provider,
		remainder.ProviderStatusNamespace,
	)
	dst.TypeMeta.APIVersion = dstApiVersion

	dst.Spec.Run.Parameters = dst.Spec.Run.RuntimeParameters
	dst.Spec.Run.RuntimeParameters = nil

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunConfiguration)
	dstApiVersion := dst.APIVersion
	remainder := RunConfigurationConversionRemainder{}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Status.ObservedPipelineVersion = src.Status.Dependencies.Pipeline.Version
	dst.Status.TriggeredPipelineVersion = src.Status.Triggers.Pipeline.Version

	dst.Spec.Run.Provider = src.Spec.Run.Provider.Name
	dst.Status.Provider = src.Status.Provider.Name
	remainder.ProviderNamespace = src.Spec.Run.Provider.Namespace
	remainder.ProviderStatusNamespace = src.Status.Provider.Namespace
	dst.Spec.Run.RuntimeParameters = dst.Spec.Run.Parameters
	dst.Spec.Run.Parameters = nil

	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Status.SynchronizationState = src.Status.Conditions.GetSyncStateFromReason()

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
