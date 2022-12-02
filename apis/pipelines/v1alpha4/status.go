package v1alpha4

import (
	"encoding/json"
	"github.com/sky-uk/kfp-operator/apis"
	"strings"
)

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Pattern:=`^[\w-]+(?::[\w-]+)?$`
type ProviderId struct {
	Provider string `json:"-"`
	Id       string `json:"-"`
}

func (pid *ProviderId) String() string {
	if pid.Provider == "" || pid.Id == "" {
		return pid.Id
	}

	return strings.Join([]string{pid.Provider, pid.Id}, ":")
}

func (pid *ProviderId) fromString(raw string) {
	providerAndId := strings.Split(raw, ":")

	if len(providerAndId) == 2 {
		pid.Provider = providerAndId[0]
		pid.Id = providerAndId[1]
	} else if len(providerAndId) == 1 {
		pid.Id = providerAndId[0]
	}
}

func (pid *ProviderId) MarshalJSON() ([]byte, error) {
	return json.Marshal(pid.String())
}

func (pid *ProviderId) UnmarshalJSON(bytes []byte) error {
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
	ProviderId           ProviderId                `json:"providerId,omitempty"`
	SynchronizationState apis.SynchronizationState `json:"synchronizationState,omitempty"`
	Version              string                    `json:"version,omitempty"`
	ObservedGeneration   int64                     `json:"observedGeneration,omitempty"`
}
