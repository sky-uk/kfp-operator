//go:build unit || decoupled || integration
// +build unit decoupled integration

package common

import "k8s.io/apimachinery/pkg/util/rand"

func RandomString() string {
	return rand.String(5)
}

func RandomInt64() int64 {
	return int64(rand.Int())
}

func RandomExceptOne() int64 {
	if n := RandomInt64(); n == 1 {
		return 2
	} else {
		return n
	}
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

func UnsafeValue[T any](t T, _ error) T {
	return t
}
