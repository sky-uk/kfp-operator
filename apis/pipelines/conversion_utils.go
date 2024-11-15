package pipelines

import (
	"encoding/json"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:generate=false
type ConversionRemainder interface {
	Empty() bool
	ConversionAnnotation() string
}

func SetConversionAnnotations(
	resource metav1.Object,
	remainders ...ConversionRemainder,
) error {
	annotations := resource.GetAnnotations()
	for _, remainder := range remainders {
		if !remainder.Empty() {
			remainderJson, err := json.Marshal(remainder)
			if err != nil {
				return err
			}
			if annotations == nil {
				annotations = map[string]string{}
			}
			annotations[remainder.ConversionAnnotation()] = string(remainderJson)
		}
	}
	resource.SetAnnotations(annotations)
	return nil
}

func GetAndUnsetConversionAnnotations(
	resource metav1.Object,
	remainders ...ConversionRemainder,
) error {
	annotations := resource.GetAnnotations()
	for _, remainder := range remainders {
		if remainderJson, hasRemainder := annotations[remainder.ConversionAnnotation()]; hasRemainder {
			err := json.Unmarshal([]byte(remainderJson), remainder)
			if err != nil {
				return err
			}
			delete(annotations, remainder.ConversionAnnotation())
		}
	}
	return nil
}

func TransformInto[S any, D any](source S, destination *D) error {
	srcStatus, err := json.Marshal(source)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(srcStatus, &destination); err != nil {
		return err
	}
	return nil
}
