package v1alpha5

type TriggerConversionRemainder struct {
	Triggers Triggers `json:"triggers,omitempty"`
}

func (rcr TriggerConversionRemainder) Empty() bool {
	return len(rcr.Triggers.Schedules) == 0 && len(rcr.Triggers.OnChange) == 0
}

func (rcr TriggerConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}
