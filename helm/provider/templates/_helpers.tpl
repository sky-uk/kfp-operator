{{/*
Remove trailing '/' characters from .Values.containerRegistry.
*/}}
{{- define "kfp-operator-provider.trimmedContainerRegistry" -}}
{{ if .Values.containerRegistry -}}
{{ (trimSuffix "/" .Values.containerRegistry) }}/
{{- else -}}

{{- end }}
{{- end }}

{{/*
Namespace for provider resources.
*/}}
{{- define "kfp-operator-provider.namespace" -}}
{{ default .Values.namespace.name .Values.provider.namespace }}
{{- end -}}

{{/*
Prefix for provider resources, allows for override with .Values.prefixOverride
*/}}
{{- define "kfp-operator-provider.resource-prefix" -}}
{{- if .Values.prefixOverride }}
{{- .Values.prefixOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Generate docker image location for the provider
*/}}
{{- define "kfp-operator-provider.image" -}}
{{ include "kfp-operator-provider.trimmedContainerRegistry" . }}kfp-operator-{{ .Values.provider.type }}-provider:{{ .Chart.AppVersion }}
{{- end }}
