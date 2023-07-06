//go:build integration || decoupled || unit
// +build integration decoupled unit

package apis

import (
	"math/rand"

	"github.com/thanhpk/randstr"
	"k8s.io/apimachinery/pkg/types"
)

func RandomLowercaseString() string {
	return randstr.String(rand.Intn(15)+5, "0123456789abcdefghijklmnopqrstuvwxyz")
}

func RandomShortHash() string {
	return randstr.String(7, "0123456789abcdef")
}

func RandomString() string {
	return randstr.String(rand.Intn(15) + 5)
}

func RandomMap() map[string]string {
	size := rand.Intn(5)

	rMap := make(map[string]string, size)
	for i := 1; i <= size; i++ {
		rMap[RandomString()] = RandomString()
	}

	return rMap
}

func RandomList[T any](gen func() T) []T {
	size := rand.Intn(5)

	rList := make([]T, size)

	for i := 0; i < size; i++ {
		rList[i] = gen()
	}

	return rList
}

func RandomNamedValue() NamedValue {
	return NamedValue{Name: RandomString(), Value: RandomString()}
}
func RandomNamedValues() []NamedValue {
	return RandomList(RandomNamedValue)
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
