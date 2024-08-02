{{/*
Common labels
*/}}
{{- define "kfp-operator.eventsourceServer.labels" -}}
app: {{ include "kfp-operator.fullname" . }}-run-completion-eventsource-server
provider: {{ .ProviderName }}
{{- end }}
