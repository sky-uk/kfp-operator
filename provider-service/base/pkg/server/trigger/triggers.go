package trigger

const (
	TriggerType            = "trigger-type"
	TriggerSource          = "trigger-source"
	TriggerSourceNamespace = "trigger-source-namespace"
)

func FromHeaders(headers map[string]string) map[string]string {
	triggers := map[string]string{}

	if triggers[TriggerType] != "" {
		triggers[TriggerType] = headers[TriggerType]
	}

	if triggers[TriggerSource] != "" {
		triggers[TriggerSource] = headers[TriggerSource]
	}

	if triggers[TriggerSourceNamespace] != "" {
		triggers[TriggerSourceNamespace] = headers[TriggerSourceNamespace]
	}

	return triggers
}
