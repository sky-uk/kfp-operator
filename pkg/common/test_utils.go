//go:build unit || decoupled || integration

package common

import (
	"k8s.io/apimachinery/pkg/util/rand"
)

func RandomString() string {
	return rand.String(5)
}

func RandomInt64() int64 {
	return int64(rand.Int())
}

func RandomArtifact() Artifact {
	return Artifact{Name: RandomString(), Location: RandomString()}
}

func RandomNamespacedName() NamespacedName {
	return NamespacedName{
		Name:      RandomString(),
		Namespace: RandomString(),
	}
}
