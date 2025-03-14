package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunSchedule) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.RunSchedule)
	dstApiVersion := dst.APIVersion

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	dst.Spec.Provider = addWorkflowNamespaceToProvider(src.Spec.Provider)
	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}

func (dst *RunSchedule) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.RunSchedule)
	dstApiVersion := dst.APIVersion

	err := pipelines.TransformInto(src, &dst)
	if err != nil {
		return err
	}

	dst.TypeMeta.APIVersion = dstApiVersion

	return nil
}
