package resource

type RunSchedule struct {
	Provider Provider
}

func (rs *RunSchedule) Type() string {
	return "runschedule"
}

func (rs *RunSchedule) Create(body []byte) (ResponseBody, error) {
	return ResponseBody{}, nil
}

func (rs *RunSchedule) Update(id string, body []byte) error {
	return nil
}

func (rs *RunSchedule) Delete(id string) error {
	return nil
}
