package v1alpha2

import (
	"github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha3"
)

var ConversionAnnotations = struct {
	V1alpha3ConversionRemainder string
}{
	V1alpha3ConversionRemainder: v1alpha3.GroupVersion.Version + "." + v1alpha3.GroupVersion.Group + "/conversions.remainder",
}
