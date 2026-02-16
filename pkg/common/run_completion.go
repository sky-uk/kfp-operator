package common

import (
	"fmt"
	"time"
)

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
	RunStartTime          *time.Time      `json:"runStartTime,omitempty"`
	RunEndTime            *time.Time      `json:"runEndTime,omitempty"`
	ServingModelArtifacts []Artifact      `json:"servingModelArtifacts"`
	Artifacts             []Artifact      `json:"artifacts"`
	Provider              NamespacedName  `json:"provider"`
}

func (sre RunCompletionEvent) String() string {
	return fmt.Sprintf("{Status:%s PipelineName:%+v RunConfigurationName:%+v RunName:%+v RunId:%s ServingModelArtifacts:%+v Artifacts:%+v Provider:%+v}",
		sre.Status, sre.PipelineName, sre.RunConfigurationName, sre.RunName, sre.RunId, sre.ServingModelArtifacts, sre.Artifacts, sre.Provider)
}

type ComponentArtifactInstance struct {
	Uri      string         `json:"uri"`
	Metadata map[string]any `json:"metadata"`
}

type ComponentArtifact struct {
	Name      string                      `json:"name"`
	Artifacts []ComponentArtifactInstance `json:"artifacts"`
}

type PipelineComponent struct {
	Name               string              `json:"name"`
	ComponentArtifacts []ComponentArtifact `json:"componentArtifacts"`
}

type RunCompletionEventData struct {
	Status       RunCompletionStatus `json:"status"`
	PipelineName NamespacedName      `json:"pipelineName"`
	// Optionally render structs until https://github.com/golang/go/issues/11939 is addressed
	RunConfigurationName  *NamespacedName     `json:"runConfigurationName,omitempty"`
	RunName               *NamespacedName     `json:"runName,omitempty"`
	RunId                 string              `json:"runId"`
	RunStartTime          *time.Time          `json:"runStartTime,omitempty"`
	RunEndTime            *time.Time          `json:"runEndTime,omitempty"`
	ServingModelArtifacts []Artifact          `json:"servingModelArtifacts"`
	PipelineComponents    []PipelineComponent `json:"pipelineComponents"`
	Provider              NamespacedName      `json:"provider"`
}

func (rced RunCompletionEventData) ToRunCompletionEvent() RunCompletionEvent {
	return RunCompletionEvent{
		Status:                rced.Status,
		PipelineName:          rced.PipelineName,
		RunConfigurationName:  rced.RunConfigurationName,
		RunName:               rced.RunName,
		RunId:                 rced.RunId,
		RunStartTime:          rced.RunStartTime,
		RunEndTime:            rced.RunEndTime,
		ServingModelArtifacts: rced.ServingModelArtifacts,
		Artifacts:             nil, // to be populated later
		Provider:              rced.Provider,
	}
}
