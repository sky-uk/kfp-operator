package provider

import (
	"errors"
	"strings"

	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

type StubProvider struct {
	logger logr.Logger
}

func New(logger logr.Logger) resource.Provider {
	return &StubProvider{logger}
}

func (p *StubProvider) CreatePipeline(
	pdw resource.PipelineDefinitionWrapper,
) (string, error) {
	if strings.EqualFold(pdw.PipelineDefinition.Name.Name, "create-pipeline-fail") {
		return "", errors.New("create pipeline failed")
	} else {
		return "create-pipeline-succeeded", nil
	}
}

func (p *StubProvider) UpdatePipeline(
	pdw resource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	if strings.EqualFold(id, "update-pipeline-fail") {
		return "", errors.New("update pipeline failed")
	} else {
		return "update-pipeline-succeeded", nil
	}
}

func (p *StubProvider) DeletePipeline(id string) error {
	if strings.EqualFold(id, "delete-pipeline-fail") {
		return errors.New("delete pipeline failed")
	} else {
		return nil
	}
}

func (p *StubProvider) CreateRun(
	rd resource.RunDefinition,
) (string, error) {
	if strings.EqualFold(rd.Name.Name, "create-run-fail") {
		return "", errors.New("create run failed")
	} else {
		return "create-run-succeeded", nil
	}
}

func (p *StubProvider) DeleteRun(id string) error {
	if strings.EqualFold(id, "delete-run-fail") {
		return errors.New("delete run failed")
	} else {
		return nil
	}
}

func (p *StubProvider) CreateRunSchedule(
	rsd resource.RunScheduleDefinition,
) (string, error) {
	if strings.EqualFold(rsd.Name.Name, "create-runschedule-fail") {
		return "", errors.New("create runschedule failed")
	} else {
		return "create-runschedule-succeeded", nil
	}
}

func (p *StubProvider) UpdateRunSchedule(
	rsd resource.RunScheduleDefinition,
	id string,
) (string, error) {
	if strings.EqualFold(rsd.Name.Name, "update-runschedule-fail") {
		return "", errors.New("update runschedule failed")
	} else {
		return "update-runschedule-succeeded", nil
	}
}

func (p *StubProvider) DeleteRunSchedule(id string) error {
	if strings.EqualFold(id, "delete-runschedule-fail") {
		return errors.New("delete runschedule failed")
	} else {
		return nil
	}
}

func (p *StubProvider) CreateExperiment(
	ed resource.ExperimentDefinition,
) (string, error) {
	if strings.EqualFold(ed.Name.Name, "create-experiment-fail") {
		return "", errors.New("create experiment failed")
	} else {
		return "create-experiment-succeeded", nil
	}
}

func (p *StubProvider) UpdateExperiment(
	ed resource.ExperimentDefinition,
	id string,
) (string, error) {
	if strings.EqualFold(ed.Name.Name, "update-experiment-fail") {
		return "", errors.New("update experiment failed")
	} else {
		return "update-experiment-succeeded", nil
	}
}

func (p *StubProvider) DeleteExperiment(id string) error {
	if strings.EqualFold(id, "delete-experiment-fail") {
		return errors.New("delete experiment failed")
	} else {
		return nil
	}
}
