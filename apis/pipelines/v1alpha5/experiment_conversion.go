package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this Experiment to the Hub version.
func (src *Experiment) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Experiment)
	dstApiVersion := dst.APIVersion
	remainder := ExperimentConversionRemainder{}

	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}
	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.Provider = convertProviderTo(remainder.Provider.Name, remainder.Provider.Namespace)
	dst.Status.Provider = hub.ProviderAndId{
		Name: convertProviderTo(src.Status.ProviderId.Provider, remainder.ProviderStatusNamespace),
		Id:   src.Status.ProviderId.Id,
	}

	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}

// ConvertFrom converts from the Hub version to this version.
func (dst *Experiment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Experiment)
	dstApiVersion := dst.APIVersion
	remainder := ExperimentConversionRemainder{}

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	remainder.Provider = src.Spec.Provider
	remainder.ProviderStatusNamespace = src.Status.Provider.Name.Namespace
	dst.Status.ProviderId = convertProviderAndIdFrom(src.Status.Provider)
	dst.TypeMeta.APIVersion = dstApiVersion

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
