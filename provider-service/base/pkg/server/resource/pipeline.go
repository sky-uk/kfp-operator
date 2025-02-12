package resource

import (
	"context"
	"encoding/json"

	"github.com/sky-uk/kfp-operator/common"
)

type Pipeline struct {
	Ctx      context.Context
	Provider PipelineProvider
}

func (*Pipeline) Type() string {
	return "pipeline"
}

func (p *Pipeline) Create(body []byte) (ResponseBody, error) {
	logger := common.LoggerFromContext(p.Ctx)
	pdw := PipelineDefinitionWrapper{}
	if err := json.Unmarshal(body, &pdw); err != nil {
		logger.Error(err, "Failed to unmarshal PipelineDefinitionWrapper while creating Pipeline")
		return ResponseBody{}, &UserError{err}
	}

	id, err := p.Provider.CreatePipeline(pdw)
	if err != nil {
		logger.Error(err, "CreatePipeline failed")
		return ResponseBody{}, err
	}
	logger.Info("CreatePipeline succeeded", "response id", id)

	return ResponseBody{
		Id: id,
	}, nil
}

func (p *Pipeline) Update(id string, body []byte) (ResponseBody, error) {
	logger := common.LoggerFromContext(p.Ctx)
	pdw := PipelineDefinitionWrapper{}
	if err := json.Unmarshal(body, &pdw); err != nil {
		logger.Error(err, "Failed to unmarshal PipelineDefinitionWrapper while updating Pipeline")
		return ResponseBody{}, &UserError{err}
	}

	respId, err := p.Provider.UpdatePipeline(pdw, id)
	if err != nil {
		logger.Error(err, "UpdatePipeline failed", "id", id)
		return ResponseBody{}, err
	}
	logger.Info("UpdatePipeline succeeded", "response id", respId)

	return ResponseBody{
		Id: respId,
	}, err
}

func (p *Pipeline) Delete(id string) error {
	logger := common.LoggerFromContext(p.Ctx)
	if err := p.Provider.DeletePipeline(id); err != nil {
		logger.Error(err, "DeletePipeline failed", "id", id)
		return err
	}
	logger.Info("DeletePipeline succeeded", "id", id)

	return nil
}
