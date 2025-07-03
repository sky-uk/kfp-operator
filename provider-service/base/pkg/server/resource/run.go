package resource

import (
	"context"
	"encoding/json"
	"github.com/sky-uk/kfp-operator/common/triggers"

	"github.com/sky-uk/kfp-operator/argo/common"
)

type Run struct {
	Provider RunProvider
}

func (*Run) Type() string {
	return "run"
}

func (r *Run) Create(ctx context.Context, body []byte, headers map[string]string) (ResponseBody, error) {
	logger := common.LoggerFromContext(ctx)
	rd := RunDefinition{}

	if err := json.Unmarshal(body, &rd); err != nil {
		logger.Error(err, "Failed to unmarshal RunDefinition while creating Run")
		return ResponseBody{}, &UserError{err}
	}

	id, err := r.Provider.CreateRun(ctx, rd, triggers.FromHeaders(headers))
	if err != nil {
		logger.Error(err, "CreateRun failed")
		return ResponseBody{}, err
	}
	logger.Info("CreateRun succeeded", "response id", id)

	return ResponseBody{
		Id: id,
	}, nil
}

func (r *Run) Update(_ context.Context, _ string, _ []byte, _ map[string]string) (ResponseBody, error) {
	return ResponseBody{}, &UnimplementedError{Method: "Update", ResourceType: r.Type()}
}

func (r *Run) Delete(ctx context.Context, id string, _ map[string]string) error {
	logger := common.LoggerFromContext(ctx)
	if err := r.Provider.DeleteRun(ctx, id); err != nil {
		logger.Error(err, "DeleteRun failed", "id", id)
		return err
	}
	logger.Info("DeleteRun succeeded", "id", id)
	return nil
}
