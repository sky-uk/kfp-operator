package resource

import "encoding/json"

type RunSchedule struct {
	Provider RunScheduleProvider
}

func (*RunSchedule) Type() string {
	return "runschedule"
}

func (rs *RunSchedule) Create(body []byte) (ResponseBody, error) {
	rsd := RunScheduleDefinition{}
	if err := json.Unmarshal(body, &rsd); err != nil {
		return ResponseBody{}, &UserError{err}
	}

	id, err := rs.Provider.CreateRunSchedule(rsd)
	if err != nil {
		return ResponseBody{}, err
	}

	return ResponseBody{
		Id: id,
	}, nil
}

func (rs *RunSchedule) Update(id string, body []byte) (ResponseBody, error) {
	rsd := RunScheduleDefinition{}
	if err := json.Unmarshal(body, &rsd); err != nil {
		return ResponseBody{}, &UserError{err}
	}

	respId, err := rs.Provider.UpdateRunSchedule(rsd, id)
	if err != nil {
		return ResponseBody{}, err
	}
	return ResponseBody{
		Id: respId,
	}, nil
}

func (rs *RunSchedule) Delete(id string) error {
	if err := rs.Provider.DeleteRunSchedule(id); err != nil {
		return err
	}
	return nil
}
