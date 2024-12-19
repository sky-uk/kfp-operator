package handler

import (
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

type Resource int

const (
	Pipeline Resource = iota
	Experiment
	RunSchedule
	Run
)

func createResource(
	w http.ResponseWriter,
	r *http.Request,
	resource Resource,
) string {
	return "foo"
}
