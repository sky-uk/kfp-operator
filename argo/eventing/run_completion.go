package eventing

import (
	"encoding/json"
	"fmt"
	"strings"
)

const RunCompletionEventName = "run-completion"

type ServingModelArtifact struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type RunCompletionStatus string

var RunCompletionStatuses = struct {
	Succeeded RunCompletionStatus
	Failed    RunCompletionStatus
} {
	Succeeded: "succeeded",
	Failed: "failed",
}

// +kubebuilder:validation:Type=string
// +kubebuilder:validation:Pattern:=`^[\w-]+/[\w-]+$`
type NamespacedName struct {
	Name      string `json:"-"`
	Namespace string `json:"-"`
}

func (nsn *NamespacedName) String() string {
	return strings.Join([]string{nsn.Name, nsn.Namespace}, "/")
}

func (nsn *NamespacedName) MarshalJSON() ([]byte, error) {
	return json.Marshal(nsn.String())
}

func (nsn *NamespacedName) UnmarshalJSON(bytes []byte) error {
	var pidStr string
	err := json.Unmarshal(bytes, &pidStr)
	if err != nil {
		return err
	}

	splits := strings.Split(pidStr, "/")

	if len(splits) < 2 {
		return fmt.Errorf("namespaced name must have name and namespace")
	}

	nsn.Name = splits[0]
	nsn.Name = splits[1]

	return nil
}

type RunCompletionEvent struct {
	Status                RunCompletionStatus    `json:"status"`
	PipelineName          string                 `json:"pipelineName"`
	RunConfigurationName  string                 `json:"runConfigurationName,omitempty"`
	Run                   NamespacedName         `json:"run,omitempty"`
	ServingModelArtifacts []ServingModelArtifact `json:"servingModelArtifacts"`
}
