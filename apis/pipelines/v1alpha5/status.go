package v1alpha5

import (
	"encoding/json"
	"github.com/sky-uk/kfp-operator/apis"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

var ConditionTypes = struct {
	SynchronizationSucceeded string
}{
	SynchronizationSucceeded: "SynchronizationSucceeded",
}

// +kubebuilder:validation:Type=string
type ProviderAndId struct {
	Provider string `json:"-"`
	Id       string `json:"-"`
}

func (pid *ProviderAndId) String() string {
	if pid.Provider == "" || pid.Id == "" {
		return pid.Id
	}

	return strings.Join([]string{pid.Provider, pid.Id}, ":")
}

func (pid *ProviderAndId) fromString(raw string) {
	providerAndId := strings.Split(raw, ":")

	if len(providerAndId) == 2 {
		pid.Provider = providerAndId[0]
		pid.Id = providerAndId[1]
	} else if len(providerAndId) == 1 {
		pid.Id = providerAndId[0]
	}
}

func (pid *ProviderAndId) MarshalJSON() ([]byte, error) {
	return json.Marshal(pid.String())
}

func (pid *ProviderAndId) UnmarshalJSON(bytes []byte) error {
	var pidStr string
	err := json.Unmarshal(bytes, &pidStr)
	if err != nil {
		return err
	}

	pid.fromString(pidStr)

	return nil
}

func ConditionStatusForSynchronizationState(state apis.SynchronizationState) v1.ConditionStatus {
	switch state {
	case apis.Succeeded, apis.Deleted:
		return v1.ConditionTrue
	case apis.Failed:
		return v1.ConditionFalse
	default:
		return v1.ConditionUnknown
	}
}

type Conditions []v1.Condition

func (conditions Conditions) SynchronizationSucceeded() v1.Condition {
	return conditions.ToMap()[ConditionTypes.SynchronizationSucceeded]
}

func (conditions Conditions) ToMap() map[string]v1.Condition {
	return pipelines.ToMap(conditions, func(condition v1.Condition) (string, v1.Condition) {
		return condition.Type, condition
	})
}

func (conditions Conditions) MergeIntoConditions(condition v1.Condition) Conditions {
	conditionsAsMap := conditions.ToMap()

	existingCondition := conditionsAsMap[condition.Type]

	if existingCondition.Reason != condition.Reason || existingCondition.Status != condition.Status || existingCondition.ObservedGeneration != condition.ObservedGeneration {
		conditionsAsMap[condition.Type] = condition
	}

	return pipelines.Values(conditionsAsMap)
}

// +kubebuilder:object:generate=true
type Status struct {
	ProviderId           ProviderAndId             `json:"providerId,omitempty"`
	SynchronizationState apis.SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string                    `json:"version,omitempty"`
	ObservedGeneration   int64                     `json:"observedGeneration,omitempty"`
	Conditions           Conditions                `json:"conditions,omitempty"`
}
