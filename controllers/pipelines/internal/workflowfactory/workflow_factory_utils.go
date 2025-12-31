package workflowfactory

import "github.com/sky-uk/kfp-operator/apis"

func NamedValuesToMap(namedValues []apis.NamedValue) map[string]string {
	m := make(map[string]string)

	for _, nv := range namedValues {
		m[nv.Name] = nv.Value
	}

	return m
}
