package common

const RunCompletionEventName = "run-completion"

type ServingModelArtifact struct {
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
	Status                RunCompletionStatus    `json:"status"`
	PipelineName          NamespacedName         `json:"pipelineName"`
	// Optionally render structs until https://github.com/golang/go/issues/11939 is addressed
	RunConfigurationName  *NamespacedName        `json:"runConfigurationName,omitempty"`
	RunName               *NamespacedName        `json:"runName,omitempty"`
	RunId                 string                 `json:"runId"`
	ServingModelArtifacts []ServingModelArtifact `json:"servingModelArtifacts"`
}
