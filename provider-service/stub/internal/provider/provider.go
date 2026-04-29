package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/pkg/common"
	. "github.com/sky-uk/kfp-operator/pkg/common/testutil/provider"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

type StubProvider struct {
	logger          logr.Logger
	operatorWebhook string
	providerName    common.NamespacedName
}

func New(logger logr.Logger, operatorWebhook string, providerName common.NamespacedName) resource.Provider {
	return &StubProvider{
		logger:          logger,
		operatorWebhook: operatorWebhook,
		providerName:    providerName,
	}
}

func (p *StubProvider) CreatePipeline(
	_ context.Context,
	pdw resource.PipelineDefinitionWrapper,
) (string, error) {
	if strings.EqualFold(pdw.PipelineDefinition.Name.Name, CreatePipelineFail) {
		return "", &CreatePipelineError{}
	} else {
		return CreatePipelineSucceeded, nil
	}
}

func (p *StubProvider) UpdatePipeline(
	_ context.Context,
	_ resource.PipelineDefinitionWrapper,
	id string,
) (string, error) {
	if strings.EqualFold(id, UpdatePipelineFail) {
		return "", &UpdatePipelineError{}
	} else {
		return UpdatePipelineSucceeded, nil
	}
}

func (p *StubProvider) DeletePipeline(_ context.Context, id string) error {
	if strings.EqualFold(id, DeletePipelineFail) {
		return &DeletePipelineError{}
	} else {
		return nil
	}
}

func (p *StubProvider) CreateRun(
	_ context.Context,
	rd base.RunDefinition,
) (string, error) {
	if strings.EqualFold(rd.Name.Name, CreateRunFail) {
		return "", &CreateRunError{}
	}

	if p.operatorWebhook != "" {
		go p.fireRunCompletionEvent(rd)
	}

	return CreateRunSucceeded, nil
}

func (p *StubProvider) fireRunCompletionEvent(rd base.RunDefinition) {
	startTime := time.Now().UTC()
	time.Sleep(5 * time.Second)
	now := time.Now().UTC()
	runId := fmt.Sprintf("stub-run-%d", now.Unix())

	event := common.RunCompletionEventData{
		Status:               common.RunCompletionStatuses.Succeeded,
		PipelineName:         rd.PipelineName,
		RunConfigurationName: rd.RunConfigurationName.NonEmptyPtr(),
		RunName:              rd.Name.NonEmptyPtr(),
		RunId:                runId,
		RunStartTime:         &startTime,
		RunEndTime:           &now,
		PipelineComponents:   []common.PipelineComponent{},
		Provider:             p.providerName,
	}

	body, err := json.Marshal(event)
	if err != nil {
		p.logger.Error(err, "failed to marshal run completion event")
		return
	}

	p.logger.Info("firing run completion event", "webhook", p.operatorWebhook, "runId", runId, "runName", rd.Name)

	resp, err := http.Post(p.operatorWebhook, "application/json", bytes.NewReader(body))
	if err != nil {
		p.logger.Error(err, "failed to send run completion event")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		p.logger.Info("run completion event sent successfully", "runId", runId)
	} else {
		p.logger.Error(fmt.Errorf("unexpected status code: %d", resp.StatusCode), "run completion event failed", "runId", runId)
	}
}

func (p *StubProvider) DeleteRun(_ context.Context, id string) error {
	if strings.EqualFold(id, DeleteRunFail) {
		return &DeleteRunError{}
	} else {
		return nil
	}
}

func (p *StubProvider) CreateRunSchedule(
	_ context.Context,
	rsd base.RunScheduleDefinition,
) (string, error) {
	if strings.EqualFold(rsd.Name.Name, CreateRunScheduleFail) {
		return "", &CreateRunScheduleError{}
	} else {
		return CreateRunScheduleSucceeded, nil
	}
}

func (p *StubProvider) UpdateRunSchedule(
	_ context.Context,
	rsd base.RunScheduleDefinition,
	_ string,
) (string, error) {
	if strings.EqualFold(rsd.Name.Name, UpdateRunScheduleFail) {
		return "", &UpdateRunScheduleError{}
	} else {
		return UpdateRunScheduleSucceeded, nil
	}
}

func (p *StubProvider) DeleteRunSchedule(_ context.Context, id string) error {
	if strings.EqualFold(id, DeleteRunScheduledFail) {
		return &DeleteRunScheduleError{}
	} else {
		return nil
	}
}

func (p *StubProvider) CreateExperiment(
	_ context.Context,
	ed base.ExperimentDefinition,
) (string, error) {
	if strings.EqualFold(ed.Name.Name, CreateExperimentFail) {
		return "", &CreateExperimentError{}
	} else {
		return CreateExperimentSucceeded, nil
	}
}

func (p *StubProvider) UpdateExperiment(
	_ context.Context,
	ed base.ExperimentDefinition,
	_ string,
) (string, error) {
	if strings.EqualFold(ed.Name.Name, UpdateExperimentFail) {
		return "", &UpdateExperimentError{}
	} else {
		return UpdateExperimentSucceeded, nil
	}
}

func (p *StubProvider) DeleteExperiment(_ context.Context, id string) error {
	if strings.EqualFold(id, DeleteExperimentFail) {
		return &DeleteExperimentError{}
	} else {
		return nil
	}
}
