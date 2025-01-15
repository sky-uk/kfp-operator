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

func (rs *Run) Create(body []byte) (ResponseBody, error) {
	rd := RunDefinition{}

	err := json.Unmarshal(body, &rd)
	if err != nil {
		return ResponseBody{}, err
	}

	id, err := rs.Provider.CreateRun(rd)
	if err != nil {
		rs.Logger.Error(err, "CreateRun failed")
		return ResponseBody{}, err
	}
	rs.Logger.Info("CreateRun succeeded", "response id", id)

	return ResponseBody{
		Id: id,
	}, nil
}

func (*Run) Update(_ string, _ []byte) error {
	return nil
}

func (rs *Run) Delete(id string) error {
	if err := rs.Provider.DeleteRun(id); err != nil {
		return err
	}
	return nil
}
