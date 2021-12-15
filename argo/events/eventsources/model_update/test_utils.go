//go:build unit || decoupled
// +build unit decoupled

package main

import "k8s.io/apimachinery/pkg/util/rand"

func randomString() string {
	return rand.String(5)
}
