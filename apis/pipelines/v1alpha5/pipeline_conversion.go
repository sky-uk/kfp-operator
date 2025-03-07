package v1alpha5

import (
	"encoding/json"
	"errors"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Pipeline)
	dstApiVersion := dst.APIVersion

	v1alpha5Remainder := hub.PipelineConversionRemainder{}
	if err := pipelines.GetAndUnsetConversionAnnotations(src, &v1alpha5Remainder); err != nil {
		return err
	}

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	dst.Spec.Provider = getProviderAnnotation(src)
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Status.Provider = convertProviderAndIdTo(src.Status.ProviderId)

	removeProviderAnnotation(dst)

	if !v1alpha5Remainder.Empty() {
		dst.Spec.Framework = &hub.PipelineFramework{
			Type:       v1alpha5Remainder.Framework.Type,
			Parameters: v1alpha5Remainder.Framework.Parameters,
		}
	} else if src.Spec.TfxComponents != "" {
		marshal, err := json.Marshal(src.Spec.TfxComponents)
		if err != nil {
			return err
		}
		dst.Spec.Framework = &hub.PipelineFramework{
			Type:       "tfx",
			Parameters: map[string]*apiextensionsv1.JSON{"components": {Raw: marshal}},
		}
	} else {
		return errors.New("missing tfx components in framework parameters")
	}

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)
	dstApiVersion := dst.APIVersion

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}
	setProviderAnnotation(src.Spec.Provider, &dst.ObjectMeta)
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Status.ProviderId = convertProviderAndIdFrom(src.Status.Provider)

	frameworkType := src.Spec.Framework.Type
	frameworkParameters := src.Spec.Framework.Parameters
	if frameworkType != "tfx" {
		v1alpha5Remainder := hub.PipelineConversionRemainder{Framework: *src.Spec.Framework}
		return pipelines.SetConversionAnnotations(dst, v1alpha5Remainder)
	} else if frameworkType == "tfx" && frameworkParameters != nil {
		components, componentsExists := frameworkParameters["components"]
		if componentsExists {
			var res string
			if err = json.Unmarshal(components.Raw, &res); err != nil {
				return err
			}
			dst.Spec.TfxComponents = res
			return nil
		} else {
			return errors.New("missing tfx components in framework parameters")
		}
	} else {
		return errors.New("missing tfx framework parameters")
	}
}
