package resource

import (
	"encoding/json"
)

type Pipeline struct {
	Provider PipelineProvider
}

func (*Pipeline) Type() string {
	return "pipeline"
}

func (p *Pipeline) Create(body []byte) (ResponseBody, error) {
	pdw := PipelineDefinitionWrapper{}
	err := json.Unmarshal(body, &pdw)
	if err != nil {
		return ResponseBody{}, err
	}

	id, err := p.Provider.CreatePipeline(pdw)
	if err != nil {
		return ResponseBody{}, err
	}

	return ResponseBody{
		Id: id,
	}, nil
}

func (p *Pipeline) Update(id string, body []byte) error {
	pdw := PipelineDefinitionWrapper{}
	if err := json.Unmarshal(body, &pdw); err != nil {
		return err
	}

	_, err := p.Provider.UpdatePipeline(pdw, id)
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
