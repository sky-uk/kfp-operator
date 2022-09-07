package v1alpha2

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

type PipelineConversionRemainder struct {
	BeamArgs []apis.NamedValue `json:"beamArgs,omitempty"`
	Env      []apis.NamedValue `json:"env,omitempty"`
}

func (pcr PipelineConversionRemainder) empty() bool {
	return pcr.BeamArgs == nil && pcr.Env == nil
}

func (src *Pipeline) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Pipeline)

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status
	dst.Spec.Env = mapToNamedValues(src.Spec.Env)
	dst.Spec.BeamArgs = mapToNamedValues(src.Spec.BeamArgs)
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents

	remainder := PipelineConversionRemainder{}
	if err := retrieveAndUnsetConversionAnnotations(dst, &remainder); err != nil {
		return err
	}

	dst.Spec.BeamArgs = append(dst.Spec.BeamArgs, remainder.BeamArgs...)
	dst.Spec.Env = append(dst.Spec.Env, remainder.Env...)

	return nil
}

func (dst *Pipeline) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Pipeline)

	remainder := PipelineConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	dst.Status = src.Status
	dst.Spec.Env, remainder.Env = namedValuesToMap(src.Spec.Env)
	dst.Spec.BeamArgs, remainder.BeamArgs = namedValuesToMap(src.Spec.BeamArgs)
	dst.Spec.Image = src.Spec.Image
	dst.Spec.TfxComponents = src.Spec.TfxComponents

	if err := setConversionAnnotations(dst, remainder); err != nil {
		return err
	}

	return nil
}
