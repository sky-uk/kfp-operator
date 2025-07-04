package triggers

import (
	"fmt"
	"regexp"
	"strings"
)

const Group = "pipelines.kubeflow.org"

const (
	Type            = "trigger-type"
	Source          = "trigger-source"
	SourceNamespace = "trigger-source-namespace"
)

const (
	OnChangePipeline = "onChangePipeline"
	OnChangeRunSpec  = "onChangeRunSpec"
	RunConfiguration = "runConfiguration"
)

var (
	TriggerByTypeLabel            = fmt.Sprintf("%s/%s", Group, Type)
	TriggerBySourceLabel          = fmt.Sprintf("%s/%s", Group, Source)
	TriggerBySourceNamespaceLabel = fmt.Sprintf("%s/%s", Group, SourceNamespace)
)

type Indicator struct {
	Type            string `json:"type"`
	Source          string `json:"source"`
	SourceNamespace string `json:"sourceNamespace"`
}

func (i Indicator) AsK8sLabels() map[string]string {
	labels := map[string]string{}
	if i.Type != "" {
		labels[TriggerByTypeLabel] = sanitise(i.Type)
	}
	if i.Source != "" {
		labels[TriggerBySourceLabel] = sanitise(i.Source)
	}
	if i.SourceNamespace != "" {
		labels[TriggerBySourceNamespaceLabel] = sanitise(i.SourceNamespace)
	}
	return labels
}

func FromLabels(labels map[string]string) Indicator {
	return Indicator{
		Type:            labels[TriggerByTypeLabel],
		Source:          labels[TriggerBySourceLabel],
		SourceNamespace: labels[TriggerBySourceNamespaceLabel],
	}
}

// sanitise removes any characters that are not alphanumeric, underscore, or hyphen.
// It also replaces slashes with underscores as run configurations maybe namespaced but / is not valid in label values.
func sanitise(s string) string {
	regex := regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
	s = strings.ReplaceAll(s, "/", "_")
	s = regex.ReplaceAllString(s, "")
	return s
}
