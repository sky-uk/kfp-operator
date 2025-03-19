package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Experiment{}

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

	dst.Spec.Provider = convertProviderTo(src.Spec.Provider, remainder.ProviderNamespace)
	dst.Status.Provider.Name = convertProviderTo(
		src.Status.Provider.Name,
		remainder.ProviderStatusNamespace,
	)
	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}

// ConvertFrom converts from the Hub version to this version.
func (dst *Experiment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Experiment)
	dstApiVersion := dst.APIVersion
	remainder := ExperimentConversionRemainder{}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.Provider = src.Spec.Provider.Name
	dst.Status.Provider.Name = src.Status.Provider.Name.Name
	remainder.ProviderNamespace = src.Spec.Provider.Namespace
	remainder.ProviderStatusNamespace = src.Status.Provider.Name.Namespace

	dst.TypeMeta.APIVersion = dstApiVersion

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
