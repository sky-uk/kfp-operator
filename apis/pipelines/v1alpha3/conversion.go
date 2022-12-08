package v1alpha3

import (
	"github.com/sky-uk/kfp-operator/apis"
)

var conversionAnnotation = GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"

type RunConfigurationConversionRemainder struct {
	RuntimeParameters []apis.NamedValue `json:"runtimeParameters,omitempty"`
}

func (rcr RunConfigurationConversionRemainder) Empty() bool {
	return rcr.RuntimeParameters == nil
}

func (rcr RunConfigurationConversionRemainder) ConversionAnnotation() string {
	return conversionAnnotation
}

type PipelineConversionRemainder struct {
	BeamArgs []apis.NamedValue `json:"beamArgs,omitempty"`
	Env      []apis.NamedValue `json:"env,omitempty"`
}

func (pcr PipelineConversionRemainder) Empty() bool {
	return pcr.BeamArgs == nil && pcr.Env == nil
}

func (pcr PipelineConversionRemainder) ConversionAnnotation() string {
	return conversionAnnotation
}
