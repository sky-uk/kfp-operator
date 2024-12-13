{{/*
Common labels
*/}}
{{- define "kfp-operator-provider.service.labels" -}}
app: {{ include "kfp-operator-provider.resource-prefix" . }}-service
provider: {{ .Values.provider.name }}
{{- end }}
