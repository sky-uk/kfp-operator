package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

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

	namespacedName := convertProviderTo(remainder.Provider)
	dst.Spec.Provider = namespacedName
	dst.Status.Provider = convertProviderAndIdTo(
		src.Status.ProviderId,
		namespacedName.Namespace,
	)
	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}

func (dst *Experiment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Experiment)
	dstApiVersion := dst.APIVersion
	remainder := ExperimentConversionRemainder{}

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	remainder.Provider = src.Spec.Provider
	dst.Status.ProviderId = convertProviderAndIdFrom(src.Status.Provider)
	dst.TypeMeta.APIVersion = dstApiVersion

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
