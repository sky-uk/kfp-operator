package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Run) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Run)
	dstApiVersion := dst.APIVersion

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	dst.Spec.Provider = getProviderAnnotation(src)
	dst.TypeMeta.APIVersion = hub.GroupVersion.String()

	removeProviderAnnotation(dst)
	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}

func (dst *Run) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Run)
	dstApiVersion := dst.APIVersion

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}
	setProviderAnnotation(src.Spec.Provider, &dst.ObjectMeta)
	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}
