//go:build unit || decoupled || integration

package testutil

import (
	"github.com/sky-uk/kfp-operator/common"
	"k8s.io/apimachinery/pkg/util/rand"
)

func RandomString() string {
	return rand.String(5)
}

func RandomNamespacedName() common.NamespacedName {
	return common.NamespacedName{
		Name:      RandomString(),
		Namespace: RandomString(),
	}
}
