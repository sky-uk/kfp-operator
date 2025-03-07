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

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	dst.Spec.Provider = getProviderAnnotation(src)
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Status.Provider = convertProviderAndIdTo(src.Status.ProviderId)

	removeProviderAnnotation(dst)

	marshal, err := json.Marshal(src.Spec.TfxComponents)
	if err != nil {
		return err
	}
	dst.Spec.Framework = &hub.PipelineFramework{
		Type:       "tfx",
		Parameters: map[string]*apiextensionsv1.JSON{"components": {Raw: marshal}},
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

	var res string
	if src.Spec.Framework.Parameters == nil {
		return errors.New("missing components in framework parameters")
	}
	if err = json.Unmarshal(src.Spec.Framework.Parameters["components"].Raw, &res); err != nil {
		return err
	}
	dst.Spec.TfxComponents = res

	return nil
}
