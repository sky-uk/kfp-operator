//go:build unit || decoupled
// +build unit decoupled

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

func RandomNamespacedName() NamespacedName {
	return NamespacedName{
		Name:      RandomString(),
		Namespace: RandomString(),
	}
}
