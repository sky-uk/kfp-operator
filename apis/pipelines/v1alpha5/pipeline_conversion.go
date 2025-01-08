package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
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
	dst.Spec.Provider.Name = getProviderAnnotation(src)
	dst.Spec.Provider.Namespace = getProviderNamespaceAnnotation(src)
	dst.Status.Provider = convertProviderAndIdTo(src.Status.ProviderId, dst.Spec.Provider.Namespace)

	removeProviderAnnotation(dst)
	removeProviderNamespaceAnnotation(dst)

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Pipeline)
	dstApiVersion := dst.APIVersion

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}
	setProviderAnnotation(src.Spec.Provider.Name, &dst.ObjectMeta)
	setProviderNamespaceAnnotation(src.Spec.Provider.Namespace, &dst.ObjectMeta)
	dst.TypeMeta.APIVersion = dstApiVersion
	dst.Status.ProviderId = convertProviderAndIdFrom(src.Status.Provider)

	return nil
}
