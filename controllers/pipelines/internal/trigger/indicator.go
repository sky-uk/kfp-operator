package trigger

import (
	"fmt"

	"github.com/sky-uk/kfp-operator/apis"
)

const (
	Type            = "trigger-type"
	Source          = "trigger-source"
	SourceNamespace = "trigger-source-namespace"
)

const (
	TriggerByLabel = apis.Group + "/triggered-by"
)

const (
	OnChangePipeline = "onChangePipeline"
	OnChangeRunSpec  = "onChangeRunSpec"
	RunConfiguration = "runConfiguration"
)

type Indicator struct {
	Type            string `json:"type"`
	Source          string `json:"source"`
	SourceNamespace string `json:"sourceNamespace"`
}

func (i *Indicator) AsHeaders() map[string]string {
	return map[string]string{
		Type:            fmt.Sprintf("%s: %s", Type, i.Type),
		Source:          fmt.Sprintf("%s: %s", Source, i.Source),
		SourceNamespace: fmt.Sprintf("%s: %s", SourceNamespace, i.SourceNamespace),
	}
}
