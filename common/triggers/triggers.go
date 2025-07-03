package triggers

import (
	"fmt"
	"github.com/sky-uk/kfp-operator/apis"
)

const (
	Type            = "trigger-type"
	Source          = "trigger-source"
	SourceNamespace = "trigger-source-namespace"
)

var (
	TriggerByTypeLabel            = fmt.Sprintf("%s/%s", apis.Group, Type)
	TriggerBySourceLabel          = fmt.Sprintf("%s/%s", apis.Group, Source)
	TriggerBySourceNamespaceLabel = fmt.Sprintf("%s/%s", apis.Group, SourceNamespace)
)

func FromHeaders(headers map[string]string) map[string]string {
	triggers := map[string]string{}

	if headers[Type] != "" {
		triggers[Type] = headers[Type]
	}

	if headers[Source] != "" {
		triggers[Source] = headers[Source]
	}

	if headers[SourceNamespace] != "" {
		triggers[SourceNamespace] = headers[SourceNamespace]
	}

	return triggers
}
