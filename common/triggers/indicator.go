package triggers

import (
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
	Schedule         = "schedule"
)

var (
	TriggerByTypeLabel            = Group + "/" + Type
	TriggerBySourceLabel          = Group + "/" + Source
	TriggerBySourceNamespaceLabel = Group + "/" + SourceNamespace
	labelRegex                    = regexp.MustCompile(`[^a-zA-Z0-9_-]+`)
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
	s = strings.ReplaceAll(s, "/", "_")
	s = labelRegex.ReplaceAllString(s, "")
	return s
}
