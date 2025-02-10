package provider

import (
	"strings"

	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	providerConst "github.com/sky-uk/kfp-operator/common/testutil/provider"
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
		return "", &providerConst.CreatePipelineError{}
	} else {
		return providerConst.CreatePipelineSuccess, nil
	}
}

func (p *StubProvider) UpdatePipeline(
	pdw resource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	if strings.EqualFold(id, "update-pipeline-fail") {
		return "", &providerConst.UpdatePipelineError{}
	} else {
		return providerConst.UpdatePipelineSucceeded, nil
	}
}

func (p *StubProvider) DeletePipeline(id string) error {
	if strings.EqualFold(id, "delete-pipeline-fail") {
		return &providerConst.DeletePipelineError{}
	} else {
		return nil
	}
}

func (p *StubProvider) CreateRun(
	rd resource.RunDefinition,
) (string, error) {
	if strings.EqualFold(rd.Name.Name, "create-run-fail") {
		return "", &providerConst.CreateRunError{}
	} else {
		return providerConst.CreateRunSucceded, nil
	}
}

func (p *StubProvider) DeleteRun(id string) error {
	if strings.EqualFold(id, "delete-run-fail") {
		return &providerConst.DeleteRunError{}
	} else {
		return nil
	}
}

func (p *StubProvider) CreateRunSchedule(
	rsd resource.RunScheduleDefinition,
) (string, error) {
	if strings.EqualFold(rsd.Name.Name, "create-runschedule-fail") {
		return "", &providerConst.CreateRunScheduleError{}
	} else {
		return providerConst.CreateRunScheduleSucceeded, nil
	}
}

func (p *StubProvider) UpdateRunSchedule(
	rsd resource.RunScheduleDefinition,
	id string,
) (string, error) {
	if strings.EqualFold(rsd.Name.Name, "update-runschedule-fail") {
		return "", &providerConst.UpdateRunScheduleError{}
	} else {
		return providerConst.UpdateRunScheduleSucceeded, nil
	}
}

func (p *StubProvider) DeleteRunSchedule(id string) error {
	if strings.EqualFold(id, "delete-runschedule-fail") {
		return &providerConst.DeleteRunScheduleError{}
	} else {
		return nil
	}
}

func (p *StubProvider) CreateExperiment(
	ed resource.ExperimentDefinition,
) (string, error) {
	if strings.EqualFold(ed.Name.Name, "create-experiment-fail") {
		return "", &providerConst.CreateExperimentError{}
	} else {
		return providerConst.CreateExperimentSucceeded, nil
	}
}

func (p *StubProvider) UpdateExperiment(
	ed resource.ExperimentDefinition,
	id string,
) (string, error) {
	if strings.EqualFold(ed.Name.Name, "update-experiment-fail") {
		return "", &providerConst.UpdateExperimentError{}
	} else {
		return providerConst.UpdateExperimentSucceeded, nil
	}
}

func (p *StubProvider) DeleteExperiment(id string) error {
	if strings.EqualFold(id, "delete-experiment-fail") {
		return &providerConst.DeleteExperimentError{}
	} else {
		return nil
	}
}
