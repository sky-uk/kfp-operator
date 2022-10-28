package v1alpha2

import (
	"encoding/json"
	"github.com/sky-uk/kfp-operator/apis"
	hub "github.com/sky-uk/kfp-operator/apis/pipelines/v1alpha4"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

func namedValuesToMap(namedValues []apis.NamedValue) (converted map[string]string, unconverted []apis.NamedValue) {
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

func mapToNamedValues(values map[string]string) []apis.NamedValue {
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
	empty() bool
}

func setConversionAnnotations(resource v1.Object, remainder ConversionRemainder) error {
	if !remainder.empty() {
		remainderJson, err := json.Marshal(remainder)

		if err != nil {
			return err
		}

		annotations := resource.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[ConversionAnnotations.V1alpha3ConversionRemainder] = string(remainderJson)

		resource.SetAnnotations(annotations)
	}

	return nil
}

func retrieveAndUnsetConversionAnnotations(resource hub.Resource, remainder ConversionRemainder) error {
	if remainderJson, hasRemainder := resource.GetAnnotations()[ConversionAnnotations.V1alpha3ConversionRemainder]; hasRemainder {
		err := json.Unmarshal([]byte(remainderJson), remainder)
		if err != nil {
			return err
		}

		delete(resource.GetAnnotations(), ConversionAnnotations.V1alpha3ConversionRemainder)
	}
	return nil
}
