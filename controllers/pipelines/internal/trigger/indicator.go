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
	TriggerByTypeLabel            = apis.Group + "/triggered-by-type"
	TriggerBySourceLabel          = apis.Group + "/triggered-by-source"
	TriggerBySourceNamespaceLabel = apis.Group + "/triggered-by-source-namespace"
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

func (i Indicator) AsHeaders() map[string]string {
	headers := map[string]string{}
	if i.Type != "" {
		headers[Type] = fmt.Sprintf("%s: %s", Type, i.Type)
	}
	if i.Source != "" {
		headers[Source] = fmt.Sprintf("%s: %s", Source, i.Source)
	}
	if i.SourceNamespace != "" {
		headers[SourceNamespace] = fmt.Sprintf("%s: %s", SourceNamespace, i.Source)
	}
	return headers
}

func FromLabels(labels map[string]string) Indicator {
	return Indicator{
		Type:            labels[TriggerByTypeLabel],
		Source:          labels[TriggerBySourceLabel],
		SourceNamespace: labels[TriggerBySourceNamespaceLabel],
	}
}
