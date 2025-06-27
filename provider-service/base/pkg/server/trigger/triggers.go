package trigger

const (
	TriggerType            = "trigger-type"
	TriggerSource          = "trigger-source"
	TriggerSourceNamespace = "trigger-source-namespace"
)

func FromHeaders(headers map[string]string) map[string]string {
	triggers := map[string]string{}
	triggers[TriggerType] = headers[TriggerType]
	triggers[TriggerSource] = headers[TriggerSource]
	triggers[TriggerSourceNamespace] = headers[TriggerSourceNamespace]
	return triggers
}
