package v1alpha5

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

	dst.Spec.Provider = getProviderAnnotation(src)
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Status.Provider = convertProviderAndIdTo(src.Status.ProviderId)

	removeProviderAnnotation(dst)

	if !remainder.Empty() {
		dst.Spec.Framework = remainder.Framework
	} else if src.Spec.TfxComponents != "" {
		framework := hub.NewPipelineFramework("tfx")
		hub.AddComponentsToFrameworkParams(src.Spec.TfxComponents, &framework)
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
	setProviderAnnotation(src.Spec.Provider, &dst.ObjectMeta)
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Status.ProviderId = convertProviderAndIdFrom(src.Status.Provider)

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
