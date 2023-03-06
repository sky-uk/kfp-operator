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
	PipelineName          string                 `json:"pipelineName"`
	RunConfigurationName  string                 `json:"runConfigurationName,omitempty"`
	RunName               NamespacedName         `json:"runName,omitempty"`
	ServingModelArtifacts []ServingModelArtifact `json:"servingModelArtifacts"`
}
