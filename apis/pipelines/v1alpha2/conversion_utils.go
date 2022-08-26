package v1alpha2

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
	"sort"
)

func namedValuesToMap(namedValues []apis.NamedValue) (map[string]string, error) {
	if len(namedValues) == 0 {
		return nil, nil
	}

	values := make(map[string]string, len(namedValues))

	for _, nv := range namedValues {
		if _, exists := values[nv.Name]; exists {
			return nil, fmt.Errorf("duplicate entry: %s", nv.Name)
		}

		values[nv.Name] = nv.Value
	}

	return values, nil
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
