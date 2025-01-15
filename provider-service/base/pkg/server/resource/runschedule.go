package resource

import "encoding/json"

type RunSchedule struct {
	Provider RunScheduleProvider
}

func (rs *RunSchedule) Type() string {
	return "runschedule"
}

func (rs *RunSchedule) Create(body []byte) (ResponseBody, error) {
	definition := RunScheduleDefinition{}
	err := json.Unmarshal(body, &definition)
	if err != nil {
		return ResponseBody{}, err
	}

	id, err := rs.Provider.CreateRunSchedule(definition)
	if err != nil {
		return ResponseBody{}, err
	}

	return ResponseBody{
		Id: id,
	}, nil
}

func (rs *RunSchedule) Update(id string, body []byte) error {
	definition := RunScheduleDefinition{}
	if err := json.Unmarshal(body, &definition); err != nil {
		return err
	}

	_, err := rs.Provider.UpdateRunSchedule(definition, id)
	if err != nil {
		return err
	}
	return nil
}

func (rs *RunSchedule) Delete(id string) error {
	if err := rs.Provider.DeleteRunSchedule(id); err != nil {
		return err
	}
	return nil
}
