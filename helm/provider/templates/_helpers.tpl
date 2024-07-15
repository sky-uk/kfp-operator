{{/*
Remove trailing '/' characters from .Values.containerRegistry.
*/}}
{{- define "kfp-operator-provider.trimmedContainerRegistry" -}}
{{ if .Values.containerRegistry -}}
{{ (trimSuffix "/" .Values.containerRegistry) }}/
{{- else -}}

{{- end }}
{{- end }}

{{- define "kfp-operator-provider.namespace" -}}
{{ default .Values.namespace.name .Values.provider.namespace }}
{{- end -}}

{{- define "kfp-operator-provider.resource-prefix" -}}
{{- if .Values.prefixOverride }}
{{- .Values.prefixOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{- define "kfp-operator-provider.image" -}}
{{ include "kfp-operator-provider.trimmedContainerRegistry" . }}kfp-operator-{{ .Values.provider.type }}-provider:{{ .Chart.AppVersion }}
{{- end }}
