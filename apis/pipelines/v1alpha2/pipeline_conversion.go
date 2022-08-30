package v1alpha2

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Pipeline)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status
	dst.Spec.Env = mapToNamedValues(src.Spec.Env)
	dst.Spec.BeamArgs = mapToNamedValues(src.Spec.BeamArgs)
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Pipeline)

	var err error

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status
	dst.Spec.Env, err = namedValuesToMap(src.Spec.Env)
	if err != nil {
		return err
	}
	dst.Spec.BeamArgs, err = namedValuesToMap(src.Spec.BeamArgs)
	if err != nil {
		return err
	}
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents
	return nil
}
