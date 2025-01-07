package utils

import "github.com/sky-uk/kfp-operator/argo/common"

func ResourceNameFromNamespacedName(namespacedName common.NamespacedName) (string, error) {
	return namespacedName.SeparatedString("-")
}
