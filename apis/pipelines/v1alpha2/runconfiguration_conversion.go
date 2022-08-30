package v1alpha2

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = mapToNamedValues(src.Spec.RuntimeParameters)
	dst.Status = v1alpha3.RunConfigurationStatus(src.Status)

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.RunConfiguration)

	var err error

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters, err = namedValuesToMap(src.Spec.RuntimeParameters)
	if err != nil {
		return err
	}
	dst.Status = RunConfigurationStatus(src.Status)

	return nil
}