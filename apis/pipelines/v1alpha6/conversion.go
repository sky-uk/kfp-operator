package v1alpha6

import (
	"github.com/sky-uk/kfp-operator/argo/common"
)

var DefaultProviderNamespace string
var DefaultTfxImage string

func convertProviderTo(
	provider string,
	remainderNamespace string,
) common.NamespacedName {
	var namespace string
	if remainderNamespace == "" && provider != "" {
		namespace = DefaultProviderNamespace
	} else {
		namespace = remainderNamespace
	}

	return common.NamespacedName{
		Name:      provider,
		Namespace: namespace,
	}
}
