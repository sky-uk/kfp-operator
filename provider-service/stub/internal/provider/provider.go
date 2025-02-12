package provider

import (
	"strings"

	"github.com/go-logr/logr"
	. "github.com/sky-uk/kfp-operator/common/testutil/provider"
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
	if strings.EqualFold(pdw.PipelineDefinition.Name.Name, CreatePipelineFail) {
		return "", &CreatePipelineError{}
	} else {
		return CreatePipelineSucceeded, nil
	}
}

func (p *StubProvider) UpdatePipeline(
	pdw resource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	if strings.EqualFold(id, UpdatePipelineFail) {
		return "", &UpdatePipelineError{}
	} else {
		return UpdatePipelineSucceeded, nil
	}
}

func (p *StubProvider) DeletePipeline(id string) error {
	if strings.EqualFold(id, DeletePipelineFail) {
		return &DeletePipelineError{}
	} else {
		return nil
	}
}

func (p *StubProvider) CreateRun(
	rd resource.RunDefinition,
) (string, error) {
	if strings.EqualFold(rd.Name.Name, CreateRunFail) {
		return "", &CreateRunError{}
	} else {
		return CreateRunSucceeded, nil
	}
}

func (p *StubProvider) DeleteRun(id string) error {
	if strings.EqualFold(id, DeleteRunFail) {
		return &DeleteRunError{}
	} else {
		return nil
	}
}

func (p *StubProvider) CreateRunSchedule(
	rsd resource.RunScheduleDefinition,
) (string, error) {
	if strings.EqualFold(rsd.Name.Name, CreateRunScheduleFail) {
		return "", &CreateRunScheduleError{}
	} else {
		return CreateRunScheduleSucceeded, nil
	}
}

func (p *StubProvider) UpdateRunSchedule(
	rsd resource.RunScheduleDefinition,
	id string,
) (string, error) {
	if strings.EqualFold(rsd.Name.Name, UpdateRunScheduleFail) {
		return "", &UpdateRunScheduleError{}
	} else {
		return UpdateRunScheduleSucceeded, nil
	}
}

func (p *StubProvider) DeleteRunSchedule(id string) error {
	if strings.EqualFold(id, DeleteRunScheduledFail) {
		return &DeleteRunScheduleError{}
	} else {
		return nil
	}
}

func (p *StubProvider) CreateExperiment(
	ed resource.ExperimentDefinition,
) (string, error) {
	if strings.EqualFold(ed.Name.Name, CreateExperimentFail) {
		return "", &CreateExperimentError{}
	} else {
		return CreateExperimentSucceeded, nil
	}
}

func (p *StubProvider) UpdateExperiment(
	ed resource.ExperimentDefinition,
	id string,
) (string, error) {
	if strings.EqualFold(ed.Name.Name, UpdateExperimentFail) {
		return "", &UpdateExperimentError{}
	} else {
		return UpdateExperimentSucceeded, nil
	}
}

func (p *StubProvider) DeleteExperiment(id string) error {
	if strings.EqualFold(id, DeleteExperimentFail) {
		return &DeleteExperimentError{}
	} else {
		return nil
	}
}
