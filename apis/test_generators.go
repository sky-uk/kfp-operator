//go:build integration || decoupled || unit
// +build integration decoupled unit

package apis

import (
	"math/rand"

	"github.com/thanhpk/randstr"
	"k8s.io/apimachinery/pkg/types"
)

func RandomLowercaseString() string {
	return randstr.String(rand.Intn(20)+1, "0123456789abcdefghijklmnopqrstuvwxyz")
}

func RandomShortHash() string {
	return randstr.String(7, "0123456789abcdef")
}

func RandomString() string {
	return randstr.String(rand.Intn(20) + 1)
}

func RandomNamedValues() []NamedValue {
	size := rand.Intn(5)

	rMap := make([]NamedValue, size)
	for i := 0; i < size; i++ {
		rMap[i] = NamedValue{Name: RandomString(), Value: RandomString()}
	}

	return rMap
}

func RandomSynchronizationState() SynchronizationState {
	synchronizationStates := []SynchronizationState{
		"",
		Creating,
		Succeeded,
		Updating,
		Deleting,
		Deleted,
		Failed,
	}

	return synchronizationStates[rand.Intn(len(synchronizationStates))]
}

func RandomNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      RandomString(),
		Namespace: RandomString(),
	}
}
