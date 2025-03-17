package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

var _ conversion.Convertible = &Experiment{}

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
	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}

func (dst *Experiment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Experiment)
	dstApiVersion := dst.APIVersion
	remainder := ExperimentConversionRemainder{}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.Provider = convertProviderFrom2(src.Spec.Provider, &remainder)
	dst.TypeMeta.APIVersion = dstApiVersion

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
