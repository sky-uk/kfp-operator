package provider

import (
	"fmt"
)

const (
	CreatePipelineSuccess   = "create-pipeline-succeeded"
	UpdatePipelineSucceeded = "update-pipeline-succeeded"

	CreateRunSucceded = "create-run-succeeded"

	CreateRunScheduleSucceeded = "create-runschedule-succeeded"
	UpdateRunScheduleSucceeded = "update-runschedule-succeeded"

	CreateExperimentSucceeded = "create-experiment-succeeded"
	UpdateExperimentSucceeded = "update-experiment-succeeded"
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
