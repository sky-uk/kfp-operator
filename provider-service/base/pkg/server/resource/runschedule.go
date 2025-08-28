package resource

import (
	"context"
	"encoding/json"

	"github.com/sky-uk/kfp-operator/internal/log"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
)

type RunSchedule struct {
	Provider RunScheduleProvider
}

func (*RunSchedule) Type() string {
	return "runschedule"
}

func (rs *RunSchedule) Create(ctx context.Context, body []byte) (base.Output, error) {
	logger := log.LoggerFromContext(ctx)
	rsd := base.RunScheduleDefinition{}
	if err := json.Unmarshal(body, &rsd); err != nil {
		logger.Error(err, "Failed to unmarshal RunScheduleDefinition while creating RunSchedule")
		return base.Output{}, &UserError{err}
	}

	id, err := rs.Provider.CreateRunSchedule(ctx, rsd)
	if err != nil {
		logger.Error(err, "CreateRunSchedule failed")
		return base.Output{}, err
	}
	logger.Info("CreateRunSchedule succeeded", "response id", id)

	return base.Output{
		Id: id,
	}, nil
}

func (rs *RunSchedule) Update(ctx context.Context, id string, body []byte) (base.Output, error) {
	logger := log.LoggerFromContext(ctx)
	rsd := base.RunScheduleDefinition{}
	if err := json.Unmarshal(body, &rsd); err != nil {
		logger.Error(err, "Failed to unmarshal RunScheduleDefinition while updating RunSchedule")
		return base.Output{}, &UserError{err}
	}

	respId, err := rs.Provider.UpdateRunSchedule(ctx, rsd, id)
	if err != nil {
		logger.Error(err, "UpdateRunSchedule failed", "id", id)
		return base.Output{}, err
	}
	logger.Info("UpdateRunSchedule succeeded", "response id", respId)

	return base.Output{
		Id: respId,
	}, nil
}

func (rs *RunSchedule) Delete(ctx context.Context, id string) error {
	logger := log.LoggerFromContext(ctx)
	if err := rs.Provider.DeleteRunSchedule(ctx, id); err != nil {
		logger.Error(err, "DeleteRunSchedule failed", "id", id)
		return err
	}
	logger.Info("DeleteRunSchedule succeeded", "id", id)

	return nil
}
