package v1alpha2

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

type RunConfigurationConversionRemainder struct {
	RuntimeParameters []apis.NamedValue `json:"runtimeParameters,omitempty"`
}

func (rcr RunConfigurationConversionRemainder) empty() bool {
	return rcr.RuntimeParameters == nil
}

func (src *RunConfiguration) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha3.RunConfiguration)

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters = mapToNamedValues(src.Spec.RuntimeParameters)
	dst.Spec.Pipeline = v1alpha3.PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status = v1alpha3.RunConfigurationStatus(src.Status)

	remainder := RunConfigurationConversionRemainder{}
	if err := retrieveAndUnsetConversionAnnotations(dst, &remainder); err != nil {
		return err
	}
	dst.Spec.RuntimeParameters = append(dst.Spec.RuntimeParameters, remainder.RuntimeParameters...)

	return nil
}

func (dst *RunConfiguration) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha3.RunConfiguration)

	remainder := RunConfigurationConversionRemainder{}

	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.RuntimeParameters, remainder.RuntimeParameters = namedValuesToMap(src.Spec.RuntimeParameters)
	dst.Spec.Pipeline = PipelineIdentifier{Name: src.Spec.Pipeline.Name, Version: src.Spec.Pipeline.Version}
	dst.Spec.Schedule = src.Spec.Schedule
	dst.Spec.ExperimentName = src.Spec.ExperimentName
	dst.Status = RunConfigurationStatus(src.Status)

	if err := setConversionAnnotations(dst, remainder); err != nil {
		return err
	}

	return nil
}
