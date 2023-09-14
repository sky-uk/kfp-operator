package v1alpha6

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
