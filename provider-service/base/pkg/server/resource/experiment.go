package resource

import (
	"context"
	"encoding/json"

	"github.com/sky-uk/kfp-operator/argo/common"
)

type Experiment struct {
	Ctx      context.Context
	Provider ExperimentProvider
}

func (*Experiment) Type() string {
	return "experiment"
}

func (e *Experiment) Create(body []byte) (ResponseBody, error) {
	logger := common.LoggerFromContext(e.Ctx)
	ed := ExperimentDefinition{}
	if err := json.Unmarshal(body, &ed); err != nil {
		logger.Error(err, "Failed to unmarshal ExperimentDefinition while creating Experiment")
		return ResponseBody{}, &UserError{err}
	}

	id, err := e.Provider.CreateExperiment(ed)
	if err != nil {
		logger.Error(err, "CreateExperiment failed")
		return ResponseBody{}, err
	}
	logger.Info("CreateExperiment succeeded", "response id", id)

	return ResponseBody{
		Id: id,
	}, nil
}

func (e *Experiment) Update(id string, body []byte) (ResponseBody, error) {
	logger := common.LoggerFromContext(e.Ctx)
	ed := ExperimentDefinition{}
	if err := json.Unmarshal(body, &ed); err != nil {
		logger.Error(err, "Failed to unmarshal ExperimentDefinition while updating Experiment")
		return ResponseBody{}, &UserError{err}
	}

	respId, err := e.Provider.UpdateExperiment(ed, id)
	if err != nil {
		logger.Error(err, "UpdateExperiment failed", "id", id)
		return ResponseBody{}, err
	}
	logger.Info("UpdateExperiment succeeded", "response id", respId)

	return ResponseBody{
		Id: respId,
	}, nil
}

func (e *Experiment) Delete(id string) error {
	logger := common.LoggerFromContext(e.Ctx)
	if err := e.Provider.DeleteExperiment(id); err != nil {
		logger.Error(err, "DeleteExperiment failed", "id", id)
		return err
	}
	logger.Info("DeleteExperiment succeeded", "id", id)

	return nil
}
