package resources

import (
	"context"
	"encoding/json"
)

type Pipeline struct {
	ctx      context.Context
	provider Provider
}

func (p *Pipeline) Name() string {
	return "pipeline"
}

func (p *Pipeline) Create(body []byte) (ResponseBody, error) {
	definition := PipelineDefinition{}
	err := json.Unmarshal(body, &definition)
	if err != nil {
		return ResponseBody{}, err
	}

	id, err := p.provider.CreatePipeline(p.ctx, definition, "")
	if err != nil {
		return ResponseBody{}, err
	}

	return ResponseBody{
		Id: id,
	}, nil
}

func (p *Pipeline) Update(id string, body []byte) error {
	definition := PipelineDefinition{}
	if err := json.Unmarshal(body, &definition); err != nil {
		return err
	}

	_, err := p.provider.UpdatePipeline(p.ctx, definition, id, "")
	if err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) Delete(id string) error {
	if err := p.provider.DeletePipeline(p.ctx, id); err != nil {
		return err
	}
	return nil
}
