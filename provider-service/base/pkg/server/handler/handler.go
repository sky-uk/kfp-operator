package handler

import (
	"errors"
	"net/http"
)

type PipelineResource struct {
	Name                 string
	Version              string
	PipelineVersion      string
	RunConfigurationName string
	ExperimentName       string
	// RuntimeParameters
	// Artifacts
}

type ExperimentResource struct {
}

type RunScheduleResource struct {
}

type RunResource struct {
}

const (
	Pipeline = "pipeline"
	Experiment = "experiment"
	RunSchedule = "runschedule"
	Run = "run"
)

func validateType(s string) bool {
	switch s {
	case Pipeline, Experiment, RunSchedule, Run:
		return true
	default:
		return false
	}
}

func createResource(
	r *http.Request,
	resource string,
) (string, error) {
	if validateType(resource) {
		return "foo", nil
	}
	return "", errors.New("placeholder")
}
