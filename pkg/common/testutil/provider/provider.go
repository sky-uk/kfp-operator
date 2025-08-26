package provider

import (
	"fmt"
)

const (
	CreatePipelineSucceeded = "create-pipeline-succeeded"
	CreatePipelineFail      = "create-pipeline-fail"
	UpdatePipelineSucceeded = "update-pipeline-succeeded"
	UpdatePipelineFail      = "update-pipeline-fail"
	DeletePipelineFail      = "delete-pipeline-fail"

	CreateRunSucceeded = "create-run-succeeded"
	CreateRunFail      = "create-run-fail"
	DeleteRunFail      = "delete-run-fail"

	CreateRunScheduleSucceeded = "create-runschedule-succeeded"
	CreateRunScheduleFail      = "create-runschedule-fail"
	UpdateRunScheduleSucceeded = "update-runschedule-succeeded"
	UpdateRunScheduleFail      = "update-runschedule-fail"
	DeleteRunScheduledFail     = "delete-runschedule-fail"

	CreateExperimentSucceeded = "create-experiment-succeeded"
	CreateExperimentFail      = "create-experiment-fail"
	UpdateExperimentSucceeded = "update-experiment-succeeded"
	UpdateExperimentFail      = "update-experiment-fail"
	DeleteExperimentFail      = "delete-experiment-fail"
)

func mkErrStr(action string, resourceType string) string {
	return fmt.Sprintf("%s %s failed", action, resourceType)
}

type CreatePipelineError struct{}

func (*CreatePipelineError) Error() string {
	return mkErrStr("create", "pipeline")
}

type UpdatePipelineError struct{}

func (*UpdatePipelineError) Error() string {
	return mkErrStr("update", "pipeline")
}

type DeletePipelineError struct{}

func (*DeletePipelineError) Error() string {
	return mkErrStr("delete", "pipeline")
}

type CreateRunError struct{}

func (*CreateRunError) Error() string {
	return mkErrStr("create", "run")
}

type DeleteRunError struct{}

func (*DeleteRunError) Error() string {
	return mkErrStr("delete", "run")
}

type CreateRunScheduleError struct{}

func (*CreateRunScheduleError) Error() string {
	return mkErrStr("create", "runschedule")
}

type UpdateRunScheduleError struct{}

func (*UpdateRunScheduleError) Error() string {
	return mkErrStr("update", "runschedule")
}

type DeleteRunScheduleError struct{}

func (*DeleteRunScheduleError) Error() string {
	return mkErrStr("delete", "runschedule")
}

type CreateExperimentError struct{}

func (*CreateExperimentError) Error() string {
	return mkErrStr("create", "experiment")
}

type UpdateExperimentError struct{}

func (*UpdateExperimentError) Error() string {
	return mkErrStr("update", "experiment")
}

type DeleteExperimentError struct{}

func (*DeleteExperimentError) Error() string {
	return mkErrStr("delete", "experiment")
}
