package resource

import (
	"context"
	"encoding/json"

	"github.com/sky-uk/kfp-operator/internal/log"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
)

type Pipeline struct {
	Provider PipelineProvider
}

func (*Pipeline) Type() string {
	return "pipeline"
}

func (p *Pipeline) Create(ctx context.Context, body []byte) (base.Output, error) {
	logger := log.LoggerFromContext(ctx)
	pdw := PipelineDefinitionWrapper{}
	if err := json.Unmarshal(body, &pdw); err != nil {
		logger.Error(err, "Failed to unmarshal PipelineDefinitionWrapper while creating Pipeline")
		return base.Output{}, &UserError{err}
	}

	id, err := p.Provider.CreatePipeline(ctx, pdw)
	if err != nil {
		logger.Error(err, "CreatePipeline failed")
		return base.Output{}, err
	}
	logger.Info("CreatePipeline succeeded", "response id", id)

	return base.Output{
		Id: id,
	}, nil
}

func (p *Pipeline) Update(ctx context.Context, id string, body []byte) (base.Output, error) {
	logger := log.LoggerFromContext(ctx)
	pdw := PipelineDefinitionWrapper{}
	if err := json.Unmarshal(body, &pdw); err != nil {
		logger.Error(err, "Failed to unmarshal PipelineDefinitionWrapper while updating Pipeline")
		return base.Output{}, &UserError{err}
	}

	respId, err := p.Provider.UpdatePipeline(ctx, pdw, id)
	if err != nil {
		logger.Error(err, "UpdatePipeline failed", "id", id)
		return base.Output{}, err
	}
	logger.Info("UpdatePipeline succeeded", "response id", respId)

	return base.Output{
		Id: respId,
	}, err
}

func (p *Pipeline) Delete(ctx context.Context, id string) error {
	logger := log.LoggerFromContext(ctx)
	if err := p.Provider.DeletePipeline(ctx, id); err != nil {
		logger.Error(err, "DeletePipeline failed", "id", id)
		return err
	}
	logger.Info("DeletePipeline succeeded", "id", id)

	return nil
}
