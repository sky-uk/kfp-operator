package v1alpha5

type ResourceConversionRemainder struct {
	Triggers []Trigger `json:"triggers,omitEmpty"`
}

func (rcr ResourceConversionRemainder) Empty() bool {
	return len(rcr.Triggers) == 0
}

func (rcr ResourceConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}
