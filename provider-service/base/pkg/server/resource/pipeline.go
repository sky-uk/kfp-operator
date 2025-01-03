package resource

import (
	"encoding/json"
)

type Pipeline struct {
	Provider Provider
}

func (p *Pipeline) Type() string {
	return "pipeline"
}

func (p *Pipeline) Create(body []byte) (ResponseBody, error) {
	definition := PipelineDefinition{}
	err := json.Unmarshal(body, &definition)
	if err != nil {
		return ResponseBody{}, err
	}

	id, err := p.Provider.CreatePipeline(definition)
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

	_, err := p.Provider.UpdatePipeline(definition, id)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pipeline) Delete(id string) error {
	if err := p.Provider.DeletePipeline(id); err != nil {
		return err
	}
	return nil
}
