package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	dst.Spec.Provider = getProviderAnnotation(src)
	removeProviderAnnotation(dst)

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}
	setProviderAnnotation(src.Spec.Provider, &dst.ObjectMeta)

	return nil
}
