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
		framework := hub.NewPipelineFramework("tfx")
		if err := hub.AddComponentsToFrameworkParams(src.Spec.TfxComponents, &framework); err != nil {
			return err
		}
		if err := hub.AddBeamArgsToFrameworkParams(src.Spec.BeamArgs, &framework); err != nil {
			return err
		}
		dst.Spec.Framework = framework

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

	if src.Spec.Framework.Type != "tfx" {
		return pipelines.SetConversionAnnotations(dst, hub.PipelineConversionRemainder{
			Framework: src.Spec.Framework,
		})
	}

	components, err := hub.ComponentsFromFramework(&src.Spec.Framework)
	if err != nil {
		return err
	}
	dst.Spec.TfxComponents = components

	beamArgs, err := hub.BeamArgsFromFramework(&src.Spec.Framework)
	if err != nil {
		return err
	}
	dst.Spec.BeamArgs = beamArgs

	return nil
}
