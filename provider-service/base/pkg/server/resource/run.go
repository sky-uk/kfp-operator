package resource

type Run struct {
	Provider Provider
}

func (rs *Run) Name() string {
	return "run"
}

func (rs *Run) Create(body []byte) (ResponseBody, error) {

	return ResponseBody{}, nil
}

func (rs *Run) Update(id string, body []byte) error {
	return nil
}

func (rs *Run) Delete(id string) error {
	return nil
}
