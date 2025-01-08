package v1beta1

import (
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/argo/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ConditionTypes = struct {
	SynchronizationSucceeded string
}{
	SynchronizationSucceeded: "Synchronized",
}

// +kubebuilder:object:generate=true
type ProviderAndId struct {
	Name common.NamespacedName `json:"name,omitempty"` //TODO: rename to NamespacedName
	Id   string                `json:"id,omitempty"`
}

func ConditionStatusForSynchronizationState(state apis.SynchronizationState) metav1.ConditionStatus {
	switch state {
	case apis.Succeeded, apis.Deleted:
		return metav1.ConditionTrue
	case apis.Failed:
		return metav1.ConditionFalse
	default:
		return metav1.ConditionUnknown
	}
}

type Conditions []metav1.Condition

func (conditions Conditions) SynchronizationSucceeded() metav1.Condition {
	return conditions.ToMap()[ConditionTypes.SynchronizationSucceeded]
}

func (conditions Conditions) ToMap() map[string]metav1.Condition {
	return pipelines.ToMap(conditions, func(condition metav1.Condition) (string, metav1.Condition) {
		return condition.Type, condition
	})
}

func (conditions Conditions) MergeIntoConditions(condition metav1.Condition) Conditions {
	conditionsAsMap := conditions.ToMap()

	existingCondition := conditionsAsMap[condition.Type]

	if existingCondition.Reason != condition.Reason || existingCondition.Status != condition.Status || existingCondition.ObservedGeneration != condition.ObservedGeneration {
		conditionsAsMap[condition.Type] = condition
	}

	return pipelines.Values(conditionsAsMap)
}

// +kubebuilder:object:generate=true
type Status struct {
	Provider             ProviderAndId             `json:"provider,omitempty"`
	SynchronizationState apis.SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string                    `json:"version,omitempty"`
	ObservedGeneration   int64                     `json:"observedGeneration,omitempty"`
	Conditions           Conditions                `json:"conditions,omitempty"`
}
