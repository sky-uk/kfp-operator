package resource

import (
	"encoding/json"
	"github.com/go-logr/logr"
)

type Run struct {
	Logger   logr.Logger
	Provider RunProvider
}

func (*Run) Type() string {
	return "run"
}

func (r *Run) Create(body []byte) (ResponseBody, error) {
	rd := RunDefinition{}

	if err := json.Unmarshal(body, &rd); err != nil {
		return ResponseBody{}, &UserError{err}
	}

	id, err := r.Provider.CreateRun(rd)
	if err != nil {
		r.Logger.Error(err, "CreateRun failed")
		return ResponseBody{}, err
	}
	r.Logger.Info("CreateRun succeeded", "response id", id)

	return ResponseBody{
		Id: id,
	}, nil
}

func (*Run) Update(_ string, _ []byte) (ResponseBody, error) {
	return ResponseBody{}, nil
}

func (r *Run) Delete(id string) error {
	if err := r.Provider.DeleteRun(id); err != nil {
		return err
	}
	return nil
}
