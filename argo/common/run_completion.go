package common

import (
	"errors"
	"fmt"
	validator "github.com/go-playground/validator/v10"
	"sync"
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

var (
	validate *validator.Validate
	mutex    sync.Mutex
)

func InitialiseValidation() (*validator.Validate, error) {
	if validate == nil {
		mutex.Lock()
		defer mutex.Unlock()
		validate = validator.New()
		err := validate.RegisterValidation("pipelineName", pipelineNameValidator)
		if err != nil {
			return nil, fmt.Errorf("failed to register pipeline name validation: %s", err)
		}
	}
	return validate, nil
}

type RunCompletionEvent struct {
	Status       RunCompletionStatus `json:"status" validate:"required"`
	PipelineName NamespacedName      `json:"pipelineName" validate:"pipelineName"`
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

func validateNamespacedName(nn *NamespacedName, key string) error {
	if nn == nil {
		return fmt.Errorf("key: %s is nil", key)
	} else if nn.Name == "" {
		return fmt.Errorf("key: %s, Name field is missing", key)
	}

	return nil
}

func (sre RunCompletionEvent) Validate(validate *validator.Validate) error {
	validateErr := validate.Struct(sre)
	runConfigurationNameValidationErr := validateNamespacedName(sre.RunConfigurationName, "RunCompletionEvent.RunConfigurationName")
	runNameValidationErr := validateNamespacedName(sre.RunName, "RunCompletionEvent.RunName")

	noValidRunNames := runConfigurationNameValidationErr != nil && runNameValidationErr != nil
	bothRunNamesPresentAndValid := runConfigurationNameValidationErr == nil && runNameValidationErr == nil

	if noValidRunNames {
		return errors.Join(runConfigurationNameValidationErr, runNameValidationErr, validateErr)
	} else if bothRunNamesPresentAndValid {
		return errors.Join(validateErr, fmt.Errorf("both RunName and RunConfigurationName are present, only one should be defined in a RunCompletionEvent"))
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
