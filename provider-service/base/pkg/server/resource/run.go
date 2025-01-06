package resource

import "encoding/json"

type Run struct {
	Provider Provider
}

func (rs *Run) Type() string {
	return "run"
}

func (rs *Run) Create(body []byte) (ResponseBody, error) {
	definition := RunDefinition{}

	err := json.Unmarshal(body, &definition)
	if err != nil {
		return ResponseBody{}, err
	}

	id, err := rs.Provider.CreateRun(definition)
	if err != nil {
		return ResponseBody{}, err
	}

	return ResponseBody{
		Id: id,
	}, nil
}

func (rs *Run) Update(_ string, _ []byte) error {
	return nil
}

func (rs *Run) Delete(id string) error {
	if err := rs.Provider.DeleteRun(id); err != nil {
		return err
	}
	return nil
}
