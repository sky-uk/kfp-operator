package resource

import (
	"context"
	"encoding/json"

	"github.com/sky-uk/kfp-operator/common"
)

type Run struct {
	Ctx      context.Context
	Provider RunProvider
}

func (*Run) Type() string {
	return "run"
}

func (r *Run) Create(body []byte) (ResponseBody, error) {
	logger := common.LoggerFromContext(r.Ctx)
	rd := RunDefinition{}

	if err := json.Unmarshal(body, &rd); err != nil {
		logger.Error(err, "Failed to unmarshal RunDefinition while creating Run")
		return ResponseBody{}, &UserError{err}
	}

	id, err := r.Provider.CreateRun(rd)
	if err != nil {
		logger.Error(err, "CreateRun failed")
		return ResponseBody{}, err
	}
	logger.Info("CreateRun succeeded", "response id", id)

	return ResponseBody{
		Id: id,
	}, nil
}

func (*Run) Update(_ string, _ []byte) (ResponseBody, error) {
	return ResponseBody{}, nil
}

func (r *Run) Delete(id string) error {
	logger := common.LoggerFromContext(r.Ctx)
	if err := r.Provider.DeleteRun(id); err != nil {
		logger.Error(err, "DeleteRun failed", "id", id)
		return err
	}
	logger.Info("DeleteRun succeeded", "id", id)
	return nil
}
