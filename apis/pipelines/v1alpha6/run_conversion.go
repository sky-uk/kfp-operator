package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Run) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Run)
	dstApiVersion := dst.APIVersion

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	providerNamespace := getProviderNamespaceAnnotation(src)
	removeProviderNamespaceAnnotation(dst)
	dst.Spec.Provider = namespaceToProvider(src.Spec.Provider, providerNamespace)
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

	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}
