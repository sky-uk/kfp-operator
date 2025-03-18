package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/argo/common"
)

var DefaultProviderNamespace string

func convertProviderTo(
	provider string,
	remainderNamespace string,
) common.NamespacedName {
	var namespace = remainderNamespace
	if namespace == "" {
		namespace = DefaultProviderNamespace
	}

	return common.NamespacedName{
		Name:      provider,
		Namespace: namespace,
	}
}
