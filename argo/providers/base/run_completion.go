package base

const RunCompletionEventName = "run-completion"

type ServingModelArtifact struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type RunCompletionStatus string

const (
	Succeeded RunCompletionStatus = "succeeded"
	Failed    RunCompletionStatus = "failed"
)

type RunCompletionEvent struct {
	Status                RunCompletionStatus    `json:"status"`
	PipelineName          string                 `json:"pipelineName"`
	RunConfigurationName  string                 `json:"runConfigurationName,omitempty"`
	ServingModelArtifacts []ServingModelArtifact `json:"servingModelArtifacts"`
}
