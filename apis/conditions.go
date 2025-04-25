package apis

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ConditionTypes = struct {
	SynchronizationSucceeded string
}{
	SynchronizationSucceeded: "Synchronized",
}

func ConditionStatusForSynchronizationState(state SynchronizationState) metav1.ConditionStatus {
	switch state {
	case Succeeded, Deleted:
		return metav1.ConditionTrue
	case Failed:
		return metav1.ConditionFalse
	default:
		return metav1.ConditionUnknown
	}
}

type Conditions []metav1.Condition

func (conditions Conditions) SynchronizationSucceeded() metav1.Condition {
	return conditions.ToMap()[ConditionTypes.SynchronizationSucceeded]
}

func (conditions Conditions) GetSyncStateFromReason() SynchronizationState {
	reason := conditions.SynchronizationSucceeded().Reason
	return SynchronisationState(reason)
}

func (conditions Conditions) SetReasonForSyncState(state SynchronizationState) Conditions {
	conditionsAsMap := conditions.ToMap()
	condition := conditionsAsMap[ConditionTypes.SynchronizationSucceeded]
	condition.Reason = string(state)
	conditionsAsMap[ConditionTypes.SynchronizationSucceeded] = condition
	return Values(conditionsAsMap)
}

// SetObservedGeneration updates all conditions that match a given type
func (conditions Conditions) SetObservedGeneration(
	conditionType string,
	generation int64,
) {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			conditions[i].ObservedGeneration = generation
		}
	}
}

func (conditions Conditions) ToMap() map[string]metav1.Condition {
	return ToMap(conditions, func(condition metav1.Condition) (string, metav1.Condition) {
		return condition.Type, condition
	})
}

func (conditions Conditions) MergeIntoConditions(condition metav1.Condition) Conditions {
	conditionsAsMap := conditions.ToMap()

	existingCondition := conditionsAsMap[condition.Type]

	if existingCondition.Reason != condition.Reason ||
		existingCondition.Status != condition.Status ||
		existingCondition.ObservedGeneration != condition.ObservedGeneration ||
		existingCondition.Message != condition.Message {
		conditionsAsMap[condition.Type] = condition
	}

	return Values(conditionsAsMap)
}

type SynchronizationState string

const (
	Creating  SynchronizationState = "Creating"
	Succeeded SynchronizationState = "Succeeded"
	Updating  SynchronizationState = "Updating"
	Deleting  SynchronizationState = "Deleting"
	Deleted   SynchronizationState = "Deleted"
	Failed    SynchronizationState = "Failed"
	Unknown   SynchronizationState = "Unknown"
)

var validStates = map[string]SynchronizationState{
	strings.ToLower(string(Creating)):  Creating,
	strings.ToLower(string(Succeeded)): Succeeded,
	strings.ToLower(string(Updating)):  Updating,
	strings.ToLower(string(Deleting)):  Deleting,
	strings.ToLower(string(Deleted)):   Deleted,
	strings.ToLower(string(Failed)):    Failed,
}

func SynchronisationState(s string) SynchronizationState {
	state, ok := validStates[strings.ToLower(s)]
	if !ok {
		state = Unknown
	}
	return state
}
