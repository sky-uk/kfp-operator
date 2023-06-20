package v1alpha5

type RunConfigurationConversionRemainder struct {
	OutputArtifactsConversionRemainder `json:",inline"`
	Triggers Triggers `json:"triggers,omitempty"`
}

func (rcr RunConfigurationConversionRemainder) Empty() bool {
	return len(rcr.Triggers.Schedules) == 0 && len(rcr.Triggers.OnChange) == 0 && rcr.OutputArtifactsConversionRemainder.Empty()
}

type OutputArtifactsConversionRemainder struct {
	Artifacts []OutputArtifact `json:"artifacts,omitempty"`
}

func (rcr OutputArtifactsConversionRemainder) Empty() bool {
	return len(rcr.Artifacts) == 0
}

func (rcr OutputArtifactsConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}