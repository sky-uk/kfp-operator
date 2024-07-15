{{/*
Common labels
*/}}
{{- define "kfp-operator-provider.eventsourceServer.labels" -}}
app: {{ include "kfp-operator-provider.resource-prefix" . }}-run-completion-eventsource-server
provider: {{ .Values.provider.name }}
{{- end }}
