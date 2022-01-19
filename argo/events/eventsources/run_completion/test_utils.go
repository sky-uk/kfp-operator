//go:build unit || decoupled
// +build unit decoupled

package run_completion

import "k8s.io/apimachinery/pkg/util/rand"

func randomString() string {
	return rand.String(5)
}

func randomInt64() int64 {
	return int64(rand.Int())
}
