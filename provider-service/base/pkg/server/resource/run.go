package resource

import (
	"context"
	"encoding/json"

	"github.com/sky-uk/kfp-operator/internal/log"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
)

type Run struct {
	Provider RunProvider
}

func (*Run) Type() string {
	return "run"
}

func (r *Run) Create(ctx context.Context, body []byte) (base.Output, error) {
	logger := log.LoggerFromContext(ctx)
	rd := base.RunDefinition{}

	if err := json.Unmarshal(body, &rd); err != nil {
		logger.Error(err, "Failed to unmarshal RunDefinition while creating Run")
		return base.Output{}, &UserError{err}
	}

	id, err := r.Provider.CreateRun(ctx, rd)
	if err != nil {
		logger.Error(err, "CreateRun failed")
		return base.Output{}, err
	}
	logger.Info("CreateRun succeeded", "response id", id)

	return base.Output{
		Id: id,
	}, nil
}

func (r *Run) Update(_ context.Context, _ string, _ []byte) (base.Output, error) {
	return base.Output{}, &UnimplementedError{Method: "Update", ResourceType: r.Type()}
}

func (r *Run) Delete(ctx context.Context, id string) error {
	logger := log.LoggerFromContext(ctx)
	if err := r.Provider.DeleteRun(ctx, id); err != nil {
		logger.Error(err, "DeleteRun failed", "id", id)
		return err
	}
	logger.Info("DeleteRun succeeded", "id", id)
	return nil
}
