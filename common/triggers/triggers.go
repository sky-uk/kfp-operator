package triggers

import "fmt"

const Group = "pipelines.kubeflow.org"

const (
	Type            = "trigger-type"
	Source          = "trigger-source"
	SourceNamespace = "trigger-source-namespace"
)

var (
	TriggerByTypeLabel            = fmt.Sprintf("%s/%s", Group, Type)
	TriggerBySourceLabel          = fmt.Sprintf("%s/%s", Group, Source)
	TriggerBySourceNamespaceLabel = fmt.Sprintf("%s/%s", Group, SourceNamespace)
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
