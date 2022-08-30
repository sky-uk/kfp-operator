package v1alpha2

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Experiment) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.Experiment)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Description = src.Spec.Description
	dst.Status = src.Status

	return nil
}

func (dst *Experiment) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.Experiment)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Description = src.Spec.Description
	dst.Status = src.Status

	return nil
}
