package v1alpha5

import (
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha6"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Provider) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*hub.Provider)

	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Image = src.Spec.Image
	dst.Spec.ExecutionMode = src.Spec.ExecutionMode
	dst.Spec.ServiceAccount = src.Spec.ServiceAccount
	dst.Spec.DefaultBeamArgs = src.Spec.DefaultBeamArgs
	dst.Spec.PipelineRootStorage = src.Spec.PipelineRootStorage
	dst.Spec.Parameters = src.Spec.Parameters
	dst.Status.Conditions = hub.Conditions(src.Status.Conditions)
	return nil
}

func (dst *Provider) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*hub.Provider)

	dst.TypeMeta = src.TypeMeta
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Image = src.Spec.Image
	dst.Spec.ExecutionMode = src.Spec.ExecutionMode
	dst.Spec.ServiceAccount = src.Spec.ServiceAccount
	dst.Spec.DefaultBeamArgs = src.Spec.DefaultBeamArgs
	dst.Spec.PipelineRootStorage = src.Spec.PipelineRootStorage
	dst.Spec.Parameters = src.Spec.Parameters
	dst.Status.Conditions = Conditions(src.Status.Conditions)
	return nil
}
