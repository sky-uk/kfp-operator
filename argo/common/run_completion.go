package common

import "fmt"

const RunCompletionEventName = "run-completion"

type Artifact struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type RunCompletionStatus string

var RunCompletionStatuses = struct {
	Succeeded RunCompletionStatus
	Failed    RunCompletionStatus
}{
	Succeeded: "succeeded",
	Failed:    "failed",
}

type RunCompletionEvent struct {
	Status       RunCompletionStatus `json:"status"`
	PipelineName NamespacedName      `json:"pipelineName"`
	// Optionally render structs until https://github.com/golang/go/issues/11939 is addressed
	RunConfigurationName  *NamespacedName `json:"runConfigurationName,omitempty"`
	RunName               *NamespacedName `json:"runName,omitempty"`
	RunId                 string          `json:"runId"`
	ServingModelArtifacts []Artifact      `json:"servingModelArtifacts"`
	Artifacts             []Artifact      `json:"artifacts,omitempty"`
	Provider              string          `json:"provider"`
}

func (sre RunCompletionEvent) String() string {
	return fmt.Sprintf("{Status:%s PipelineName:%+v RunConfigurationName:%+v RunName:%+v RunId:%s ServingModelArtifacts:%+v Artifacts:%+v Provider:%s}",
		sre.Status, sre.PipelineName, sre.RunConfigurationName, sre.RunName, sre.RunId, sre.ServingModelArtifacts, sre.Artifacts, sre.Provider)
}
