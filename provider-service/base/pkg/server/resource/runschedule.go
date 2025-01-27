package resource

import (
	"context"
	"encoding/json"

	"github.com/sky-uk/kfp-operator/argo/common"
)

type RunSchedule struct {
	Ctx      context.Context
	Provider RunScheduleProvider
}

func (*RunSchedule) Type() string {
	return "runschedule"
}

func (rs *RunSchedule) Create(body []byte) (ResponseBody, error) {
	logger := common.LoggerFromContext(rs.Ctx)
	rsd := RunScheduleDefinition{}
	if err := json.Unmarshal(body, &rsd); err != nil {
		logger.Error(err, "Create failed to unmarshal RunScheduleDefinition")
		return ResponseBody{}, &UserError{err}
	}

	id, err := rs.Provider.CreateRunSchedule(rsd)
	if err != nil {
		logger.Error(err, "CreateRunSchedule failed")
		return ResponseBody{}, err
	}
	logger.Info("CreateRunSchedule succeeded", "response id", id)

	return ResponseBody{
		Id: id,
	}, nil
}

func (rs *RunSchedule) Update(id string, body []byte) (ResponseBody, error) {
	logger := common.LoggerFromContext(rs.Ctx)
	rsd := RunScheduleDefinition{}
	if err := json.Unmarshal(body, &rsd); err != nil {
		logger.Error(err, "Update failed to unmarshal RunScheduleDefinition")
		return ResponseBody{}, &UserError{err}
	}

	respId, err := rs.Provider.UpdateRunSchedule(rsd, id)
	if err != nil {
		logger.Error(err, "UpdateRunSchedule failed", "id", id)
		return ResponseBody{}, err
	}
	logger.Info("UpdateRunSchedule succeeded", "response id", respId)

	return ResponseBody{
		Id: respId,
	}, nil
}

func (rs *RunSchedule) Delete(id string) error {
	logger := common.LoggerFromContext(rs.Ctx)
	if err := rs.Provider.DeleteRunSchedule(id); err != nil {
		logger.Error(err, "DeleteRunSchedule failed", "id", id)
		return err
	}
	logger.Info("DeleteRunSchedule succeeded", "id", id)

	return nil
}
