package workflowfactory

import "github.com/sky-uk/kfp-operator/apis"

func NamedValuesToMap(namedValues []apis.NamedValue) map[string]string {
	m := make(map[string]string)

	for _, nv := range namedValues {
		m[nv.Name] = nv.Value
	}

	return m
}

func NamedValuesToMultiMap(namedValues []apis.NamedValue) map[string][]string {
	multimap := make(map[string][]string)

	for _, nv := range namedValues {
		if _, found := multimap[nv.Name]; !found {
			multimap[nv.Name] = []string{}
		}

		multimap[nv.Name] = append(multimap[nv.Name], nv.Value)
	}

	return multimap
}
