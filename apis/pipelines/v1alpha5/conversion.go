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
	Artifacts           []OutputArtifact     `json:"artifacts,omitempty"`
	ValueFromParameters map[string]ValueFrom `json:"valueFromParameters,omitempty"`
}

func (rcr RunConversionRemainder) Empty() bool {
	return len(rcr.Artifacts) == 0
}

func (rcr RunConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}

func SplitRunTimeParameters(rts []RuntimeParameter) (namedValues []apis.NamedValue, valueFroms map[string]ValueFrom) {
	valueFroms = make(map[string]ValueFrom)

	for _, rt := range rts {
		if rt.Value != "" {
			namedValues = append(namedValues, apis.NamedValue{
				Name:  rt.Name,
				Value: rt.Value,
			})
		} else {
			valueFroms[rt.Name] = rt.ValueFrom
		}
	}

	return
}

func MergeRuntimeParameters(namedValues []apis.NamedValue, valueFroms map[string]ValueFrom) (rts []RuntimeParameter) {
	for _, namedValue := range namedValues {
		rts = append(rts, RuntimeParameter{
			Name:  namedValue.Name,
			Value: namedValue.Value,
		})
	}

	for name, valueFrom := range valueFroms {
		rts = append(rts, RuntimeParameter{
			Name:      name,
			ValueFrom: valueFrom,
		})
	}

	return
}
