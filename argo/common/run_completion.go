package common

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
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
	PipelineName NamespacedName      `json:"pipelineName" validate:"pipelineName"`
	// Optionally render structs until https://github.com/golang/go/issues/11939 is addressed
	RunConfigurationName  *NamespacedName `json:"runConfigurationName"`
	RunName               *NamespacedName `json:"runName"`
	RunId                 string          `json:"runId" validate:"required"`
	ServingModelArtifacts []Artifact      `json:"servingModelArtifacts"`
	Artifacts             []Artifact      `json:"artifacts,omitempty"`
	Provider              string          `json:"provider" validate:"required"`
}

func (sre RunCompletionEvent) String() string {
	return fmt.Sprintf("{Status:%s PipelineName:%+v RunConfigurationName:%+v RunName:%+v RunId:%s ServingModelArtifacts:%+v Artifacts:%+v Provider:%s}",
		sre.Status, sre.PipelineName, sre.RunConfigurationName, sre.RunName, sre.RunId, sre.ServingModelArtifacts, sre.Artifacts, sre.Provider)
}

func validateNamespacedName(nn *NamespacedName, key string) error {
	if nn == nil {
		return fmt.Errorf("key: %s is nil", key)
	} else if nn.Name == "" {
		return fmt.Errorf("key: %s, Name field is missing", key)
	}

	return nil
}

func (sre RunCompletionEvent) Validate() error {
	validate := validator.New()
	validate.RegisterValidation("pipelineName", pipelineNameValidator)
	validateErr := validate.Struct(sre)

	runConfigurationNameValidationErr := validateNamespacedName(sre.RunConfigurationName, "RunCompletionEvent.RunConfigurationName")
	runNameValidationErr := validateNamespacedName(sre.RunName, "RunCompletionEvent.RunName")

	if runConfigurationNameValidationErr != nil && runNameValidationErr != nil {
		return errors.Join(runConfigurationNameValidationErr, runNameValidationErr, validateErr)
	}

	return validateErr
}

func pipelineNameValidator(fl validator.FieldLevel) bool {
	nn, ok := fl.Field().Interface().(NamespacedName)
	if !ok {
		return false
	}

	return nn.Name != "" && nn.Namespace != ""
}
