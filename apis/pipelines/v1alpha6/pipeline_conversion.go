package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)
	dstApiVersion := dst.APIVersion

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)
	dstApiVersion := dst.APIVersion

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}
