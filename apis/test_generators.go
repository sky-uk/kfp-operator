//go:build integration || decoupled || unit

package apis

import (
	"github.com/sky-uk/kfp-operator/argo/common"
	"math/rand"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

func RandomNonEmptyList[T any](gen func() T) []T {
	size := rand.Intn(5) + 1

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

func RandomOf[T any](ts []T) T {
	return ts[rand.Intn(len(ts))]
}

func RandomConditionStatus() metav1.ConditionStatus {
	return RandomOf([]metav1.ConditionStatus{metav1.ConditionTrue, metav1.ConditionFalse, metav1.ConditionUnknown})
}

func RandomSynchronizationStateCondition(
	state SynchronizationState,
) metav1.Condition {
	condition := metav1.Condition{
		Type:               ConditionTypes.SynchronizationSucceeded,
		Message:            RandomString(),
		ObservedGeneration: common.RandomInt64(),
		Reason:             string(state),
		LastTransitionTime: metav1.Now().Rfc3339Copy(),
		Status:             ConditionStatusForSynchronizationState(state),
	}
	return condition
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
		Unknown,
	}

	return synchronizationStates[rand.Intn(len(synchronizationStates))]
}

func RandomNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      RandomString(),
		Namespace: RandomString(),
	}
}
