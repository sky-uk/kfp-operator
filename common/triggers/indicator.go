package triggers

import (
	"fmt"
	"regexp"
	"strings"
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
		headers[SourceNamespace] = fmt.Sprintf("%s: %s", SourceNamespace, i.SourceNamespace)
	}
	return headers
}

func (i Indicator) AsLabels() map[string]string {
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
