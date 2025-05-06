package v1alpha5

import (
	"errors"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

const storedVersionHashAnnotation = "pipelines.kubeflow.org/stored-version-hash"

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)
	dstApiVersion := dst.APIVersion
	remainder := PipelineConversionRemainder{}

	if err := pipelines.GetAndUnsetConversionAnnotations(src, &remainder); err != nil {
		return err
	}
	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.Spec.Provider = convertProviderTo(
		getProviderAnnotation(src),
		remainder.ProviderNamespace,
	)
	dst.Status.Provider = convertProviderAndIdTo(
		src.Status.ProviderId,
		remainder.ProviderStatusNamespace,
	)
	removeProviderAnnotation(dst)

	dst.TypeMeta.APIVersion = dstApiVersion

	if remainder.Framework.Type != "" {
		dst.Spec.Framework = remainder.Framework
	} else if src.Spec.TfxComponents != "" {
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

	if dst.Annotations == nil {
		dst.Annotations = map[string]string{}
	}
	dst.Annotations[storedVersionHashAnnotation] = src.Status.Version

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)
	dstApiVersion := dst.APIVersion
	remainder := PipelineConversionRemainder{}

	status := src.Status.Conditions.GetSyncStateFromReason()

	if err := pipelines.TransformInto(src, &dst); err != nil {
		return err
	}

	dst.TypeMeta.APIVersion = dstApiVersion

	setProviderAnnotation(src.Spec.Provider.Name, &dst.ObjectMeta)
	remainder.ProviderNamespace = src.Spec.Provider.Namespace
	remainder.ProviderStatusNamespace = src.Status.Provider.Name.Namespace
	dst.Status.ProviderId = convertProviderAndIdFrom(src.Status.Provider)
	dst.Status.SynchronizationState = status

	if src.Spec.Framework.Type != "tfx" {
		remainder.Framework = src.Spec.Framework
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

	return pipelines.SetConversionAnnotations(dst, &remainder)
}
