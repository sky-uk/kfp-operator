package v1alpha4

type ResourceConversionRemainder struct {
	Provider string `json:"provider,omitEmpty"`
}

func (rcr ResourceConversionRemainder) Empty() bool {
	return rcr.Provider == ""
}

func (rcr ResourceConversionRemainder) ConversionAnnotation() string {
	return GroupVersion.Version + "." + GroupVersion.Group + "/conversions.remainder"
}
