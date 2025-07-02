package trigger

const (
	TriggerType            = "trigger-type"
	TriggerSource          = "trigger-source"
	TriggerSourceNamespace = "trigger-source-namespace"
)

func FromHeaders(headers map[string]string) map[string]string {
	triggers := map[string]string{}

	if headers[TriggerType] != "" {
		triggers[TriggerType] = headers[TriggerType]
	}

	if headers[TriggerSource] != "" {
		triggers[TriggerSource] = headers[TriggerSource]
	}

	if headers[TriggerSourceNamespace] != "" {
		triggers[TriggerSourceNamespace] = headers[TriggerSourceNamespace]
	}

	return triggers
}
