package resources

type Experiment struct{}

func (e *Experiment) Name() string {
	return "experiment"
}

func (e *Experiment) Create(body []byte) (ResponseBody, error) {
	return ResponseBody{}, nil
}

func (e *Experiment) Update(id string, body []byte) error {
	return nil
}

func (e *Experiment) Delete(id string) error {
	return nil
}
