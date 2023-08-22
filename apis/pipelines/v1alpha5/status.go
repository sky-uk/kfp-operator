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

// +kubebuilder:object:generate=true
type Status struct {
	ProviderId           ProviderAndId             `json:"providerId,omitempty"`
	SynchronizationState apis.SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string                    `json:"version,omitempty"`
	ObservedGeneration   int64                     `json:"observedGeneration,omitempty"`
	Conditions           []v1.Condition            `json:"conditions,omitempty"`
}

func (st Status) WithCondition(condition v1.Condition) Status {
	conditionsAsMap := pipelines.ToMap(st.Conditions, func(condition v1.Condition) (string, v1.Condition) {
		return condition.Type, condition
	})

	existingCondition := conditionsAsMap[condition.Type]

	if existingCondition.Reason != condition.Reason || existingCondition.Status != condition.Status || existingCondition.ObservedGeneration != condition.ObservedGeneration {
		conditionsAsMap[condition.Type] = condition
		st.Conditions = pipelines.Values(conditionsAsMap)
	}

	return st
}
