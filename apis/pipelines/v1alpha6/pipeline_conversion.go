package v1alpha6

import (
	"errors"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)
	dstApiVersion := dst.APIVersion

	remainder := hub.PipelineConversionRemainder{}
	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.TypeMeta.APIVersion = dstApiVersion

	tfxComponents := src.Spec.TfxComponents
	if !remainder.Empty() {
		dst.Spec.Framework = remainder.Framework
	} else if tfxComponents != "" {
		dst.Spec.Framework = hub.ToTFXPipelineFramework(tfxComponents)
	} else {
		return errors.New("missing tfx components in framework parameters")
	}

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)
	dstApiVersion := dst.APIVersion

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.TypeMeta.APIVersion = dstApiVersion

	tfxComponents, remainder, err := hub.FromPipelineFramework(src.Spec.Framework)
	if err != nil {
		return err
	} else if tfxComponents != "" {
		dst.Spec.TfxComponents = tfxComponents
		return nil
	} else if remainder != nil {
		return pipelines.SetConversionAnnotations(dst, *remainder)
	} else {
		return errors.New("failed to process framework")
	}
}
