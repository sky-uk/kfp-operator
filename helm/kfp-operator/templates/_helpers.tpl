{{/*
Expand the name of the chart.
*/}}
{{- define "kfp-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kfp-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "kfp-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kfp-operator.labels" -}}
control-plane: controller-manager
{{- end }}

{{/*
Remove trailing '/' characters from .Values.containerRegistry.
*/}}
{{- define "kfp-operator.trimmedContainerRegistry" -}}
{{ if .Values.containerRegistry -}}
{{ (trimSuffix "/" .Values.containerRegistry) }}/
{{- else -}}

{{- end }}
{{- end }}

{{- define "kfp-operator.providerImage" -}}
{{ include "kfp-operator.trimmedContainerRegistry" . }}kfp-operator-{{ .Values.manager.provider.type }}-provider:{{ .Chart.AppVersion }}
{{- end }}

{{/*
Only render object if not empty.
*/}}
{{- define "kfp-operator.notEmptyYaml" -}}
{{- if . }}
{{ . | toYaml }}
{{- end }}
{{- end }}

{{/*
Populate configuration with fallbacks and overrides.
*/}}
{{- define "fallbackConfiguration" -}}
workflowNamespace: {{ .Values.namespace.name }}
{{- end }}
{{- define "configurationOverrides" -}}
providerConfigFile: provider.yaml
workflowTemplatePrefix: {{ include "kfp-operator.fullname" . }}-
{{- if .Values.manager.multiversion.enabled }}
multiversion: true
{{- end -}}
{{- end }}

{{- define "kfp-operator.configuration" -}}
{{ merge (include "configurationOverrides" . | fromYaml) .Values.manager.configuration (include "fallbackConfiguration" . | fromYaml) | toYaml }}
{{- end }}
