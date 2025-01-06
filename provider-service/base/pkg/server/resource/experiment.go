package resource

import "encoding/json"

type Experiment struct {
	Provider Provider
}

func (e *Experiment) Type() string {
	return "experiment"
}

func (e *Experiment) Create(body []byte) (ResponseBody, error) {
	definition := ExperimentDefinition{}
	err := json.Unmarshal(body, &definition)
	if err != nil {
		return ResponseBody{}, err
	}

	id, err := e.Provider.CreateExperiment(definition)
	if err != nil {
		return ResponseBody{}, err
	}

	return ResponseBody{
		Id: id,
	}, nil
}

func (e *Experiment) Update(id string, body []byte) error {
	definition := ExperimentDefinition{}
	if err := json.Unmarshal(body, &definition); err != nil {
		return err
	}

	_, err := e.Provider.UpdateExperiment(definition, id)
	if err != nil {
		return err
	}
	return nil
}

func (e *Experiment) Delete(id string) error {
	if err := e.Provider.DeleteExperiment(id); err != nil {
		return err
	}
	return nil
}
