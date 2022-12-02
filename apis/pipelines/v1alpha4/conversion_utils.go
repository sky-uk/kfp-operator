package v1alpha4

import (
	"encoding/json"
	"github.com/sky-uk/kfp-operator/apis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

func NamedValuesToMap(namedValues []apis.NamedValue) (converted map[string]string, unconverted []apis.NamedValue) {
	if len(namedValues) == 0 {
		return nil, nil
	}

	converted = make(map[string]string, len(namedValues))

	for _, nv := range namedValues {
		if _, exists := converted[nv.Name]; exists {
			unconverted = append(unconverted, nv)
		} else {
			converted[nv.Name] = nv.Value
		}
	}

	return
}

func MapToNamedValues(values map[string]string) []apis.NamedValue {
	var namedValues []apis.NamedValue

	for k, v := range values {
		namedValues = append(namedValues, apis.NamedValue{
			Name: k, Value: v,
		})
	}

	sort.Slice(namedValues, func(i, j int) bool {
		if namedValues[i].Name != namedValues[j].Name {
			return namedValues[i].Name < namedValues[j].Name
		} else {
			return namedValues[i].Value < namedValues[j].Value
		}
	})

	return namedValues
}

// +kubebuilder:object:generate=false
type ConversionRemainder interface {
	Empty() bool
	ConversionAnnotation() string
}

func SetConversionAnnotations(resource metav1.Object, remainders ...ConversionRemainder) error {
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

func RetrieveAndUnsetConversionAnnotations(resource metav1.Object, remainders ...ConversionRemainder) error {
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
