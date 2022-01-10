{{/*
Common labels
*/}}
{{- define "kfp-operator.eventsourceServer.labels" -}}
app: {{ include "kfp-operator.fullname" . }}-model-update-eventsource-server
{{- end }}
