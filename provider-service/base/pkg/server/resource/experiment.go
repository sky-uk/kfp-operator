package resource

import "encoding/json"

type Experiment struct {
	Provider ExperimentProvider
}

func (*Experiment) Type() string {
	return "experiment"
}

func (e *Experiment) Create(body []byte) (ResponseBody, error) {
	ed := ExperimentDefinition{}
	if err := json.Unmarshal(body, &ed); err != nil {
		return ResponseBody{}, &UserError{err}
	}

	id, err := e.Provider.CreateExperiment(ed)
	if err != nil {
		return ResponseBody{}, err
	}

	return ResponseBody{
		Id: id,
	}, nil
}

func (e *Experiment) Update(id string, body []byte) (ResponseBody, error) {
	ed := ExperimentDefinition{}
	if err := json.Unmarshal(body, &ed); err != nil {
		return ResponseBody{}, &UserError{err}
	}

	respId, err := e.Provider.UpdateExperiment(ed, id)
	if err != nil {
		return ResponseBody{}, err
	}
	return ResponseBody{
		Id: respId,
	}, nil
}

func (e *Experiment) Delete(id string) error {
	if err := e.Provider.DeleteExperiment(id); err != nil {
		return err
	}
	return nil
}
