package common

import (
	"fmt"

	validator "github.com/go-playground/validator/v10"
)

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
	Status       RunCompletionStatus `json:"status" validate:"required"`
	PipelineName NamespacedName      `json:"pipelineName"`
	// Optionally render structs until https://github.com/golang/go/issues/11939 is addressed
	RunConfigurationName  *NamespacedName `json:"runConfigurationName,omitempty"`
	RunName               *NamespacedName `json:"runName,omitempty"`
	RunId                 string          `json:"runId" validate:"required"`
	ServingModelArtifacts []Artifact      `json:"servingModelArtifacts"`
	Artifacts             []Artifact      `json:"artifacts,omitempty"`
	Provider              string          `json:"provider" validate:"required"`
}

func (sre RunCompletionEvent) String() string {
	return fmt.Sprintf("{Status:%s PipelineName:%+v RunConfigurationName:%+v RunName:%+v RunId:%s ServingModelArtifacts:%+v Artifacts:%+v Provider:%s}",
		sre.Status, sre.PipelineName, sre.RunConfigurationName, sre.RunName, sre.RunId, sre.ServingModelArtifacts, sre.Artifacts, sre.Provider)
}

func RunCompletionEventValidation(sl validator.StructLevel) {
	runCompletionEvent := sl.Current().Interface().(RunCompletionEvent)

	pipelineNamespacedNameValid := namespacedNameValidation(runCompletionEvent.PipelineName, false)
	if !pipelineNamespacedNameValid {
		sl.ReportError(NamespacedName{}, "pipelineName", "PipelineName", "run_completion_event_namespaced_name", "")
	}

	runConfigurationNameValid := false
	runConfigurationNamespacedNameNonEmpty := runCompletionEvent.RunConfigurationName != nil
	if runConfigurationNamespacedNameNonEmpty {
		runConfigurationNameValid = namespacedNameValidation(*runCompletionEvent.RunConfigurationName, true)
	}

	runNameValid := false
	runNameNamespacedNonEmpty := runCompletionEvent.RunName != nil
	if runNameNamespacedNonEmpty {
		runNameValid = namespacedNameValidation(*runCompletionEvent.RunName, true)
	}

	if !runNameValid && !runConfigurationNameValid {
		if !runNameValid {
			sl.ReportError(NamespacedName{}, "runConfigurationName", "RunConfigurationName", "run_completion_event_namespaced_name", "")
		}
		if !runConfigurationNameValid {
			sl.ReportError(NamespacedName{}, "runName", "RunName", "run_completion_event_namespaced_name", "")
		}
	}

	if runNameValid && runConfigurationNameValid {
		sl.ReportError(NamespacedName{}, "runConfigurationName", "RunConfigurationName", "run_completion_event_ambiguous_name", "")
		sl.ReportError(NamespacedName{}, "runName", "RunName", "run_completion_event_ambiguous_name", "")
	}
}

func namespacedNameValidation(nn NamespacedName, allowEmptyNamespace bool) bool {
	namespaceCheck := allowEmptyNamespace || nn.Namespace != ""
	return nn.Name != "" && namespaceCheck
}
