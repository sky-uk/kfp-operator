//go:build integration || decoupled || unit
// +build integration decoupled unit

package apis

import (
	"k8s.io/apimachinery/pkg/runtime"
	"math/rand"

	"github.com/thanhpk/randstr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func RandomMap() map[string]string {
	size := rand.Intn(5)

	rMap := make(map[string]string, size)
	for i := 1; i <= size; i++ {
		rMap[RandomString()] = RandomString()
	}

	return rMap
}

func RandomNamedValues() []NamedValue {
	size := rand.Intn(5)

	rMap := make([]NamedValue, size)
	for i := 0; i < size; i++ {
		rMap[i] = NamedValue{Name: RandomString(), Value: RandomString()}
	}

	return rMap
}

type TestResource struct {
	metav1.TypeMeta
	metav1.ObjectMeta
	NamespacedName types.NamespacedName

	Kind   string
	Status Status
}

func (tr *TestResource) GetStatus() Status {
	return tr.GetStatus()
}

func (tr *TestResource) SetStatus(status Status) {
	tr.Status = status
}

func (tr *TestResource) DeepCopyObject() runtime.Object {
	return tr
}

func (tr *TestResource) GetNamespacedName() types.NamespacedName {
	return tr.NamespacedName
}

func (tr *TestResource) GetKind() string {
	return tr.Kind
}

func RandomResource() Resource {
	return &TestResource{
		Status:         RandomStatus(),
		NamespacedName: RandomNamespacedName(),
		Kind:           RandomString(),
	}
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

func RandomStatus() Status {
	return Status{
		SynchronizationState: RandomSynchronizationState(),
		Version:              RandomString(),
		KfpId:                RandomString(),
		ObservedGeneration:   rand.Int63(),
	}
}

func RandomNamespacedName() types.NamespacedName {
	return types.NamespacedName{
		Name:      RandomString(),
		Namespace: RandomString(),
	}
}
