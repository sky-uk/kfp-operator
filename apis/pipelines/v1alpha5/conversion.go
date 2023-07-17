package v1alpha5

import (
	"github.com/sky-uk/kfp-operator/apis"
)

type RunConfigurationConversionRemainder struct {
	RunConversionRemainder `json:",inline"`
	Triggers               Triggers `json:"triggers,omitempty"`
}

func (rcr RunConfigurationConversionRemainder) Empty() bool {
	return len(rcr.Triggers.Schedules) == 0 && len(rcr.Triggers.OnChange) == 0 && rcr.RunConversionRemainder.Empty()
}

type RunConversionRemainder struct {
	Artifacts           []OutputArtifact   `json:"artifacts,omitempty"`
	ValueFromParameters []RuntimeParameter `json:"valueFromParameters,omitempty"`
}

func (rcr RunConversionRemainder) Empty() bool {
	return len(rcr.Artifacts) == 0 && len(rcr.ValueFromParameters) == 0
}

func (rcr RunConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}

func SplitRunTimeParameters(rts []RuntimeParameter) (namedValues []apis.NamedValue, valueFroms []RuntimeParameter) {
	for _, rt := range rts {
		if rt.ValueFrom != nil {
			valueFroms = append(valueFroms, rt)
		} else {
			namedValues = append(namedValues, apis.NamedValue{
				Name:  rt.Name,
				Value: rt.Value,
			})
		}
	}

	return
}

func MergeRuntimeParameters(namedValues []apis.NamedValue, valueFroms []RuntimeParameter) (rts []RuntimeParameter) {
	for _, namedValue := range namedValues {
		rts = append(rts, RuntimeParameter{
			Name:  namedValue.Name,
			Value: namedValue.Value,
		})
	}

	rts = append(rts, valueFroms...)

	return
}
