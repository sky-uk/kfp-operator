package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/argo/common"
)

var DefaultProviderNamespace string

func addWorkflowNamespaceToProvider(provider string) common.NamespacedName {
	namespacedName, err := common.NamespacedNameFromString(provider)

	if err != nil {
		panic(err)
	}

	if namespacedName.Namespace != "" {
		return namespacedName
	} else {
		return common.NamespacedName{
			Name:      namespacedName.Name,
			Namespace: DefaultProviderNamespace,
		}
	}
}
