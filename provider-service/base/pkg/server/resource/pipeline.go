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
	if err := json.Unmarshal(body, &pdw); err != nil {
		return ResponseBody{}, &UserError{err}
	}

	id, err := p.Provider.CreatePipeline(pdw)
	if err != nil {
		return ResponseBody{}, err
	}

	return ResponseBody{
		Id: id,
	}, nil
}

func (p *Pipeline) Update(id string, body []byte) (ResponseBody, error) {
	pdw := PipelineDefinitionWrapper{}
	if err := json.Unmarshal(body, &pdw); err != nil {
		return ResponseBody{}, &UserError{err}
	}

	respId, err := p.Provider.UpdatePipeline(pdw, id)
	if err != nil {
		return ResponseBody{}, err
	}
	return ResponseBody{
		Id: respId,
	}, err
}

func (p *Pipeline) Delete(id string) error {
	if err := p.Provider.DeletePipeline(id); err != nil {
		return err
	}
	return nil
}
