{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "kfp-operator.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := .Chart.Name }}
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

{{- define "kfp-operator.compilerImage" -}}
{{ if .Values.manager.argo.compilerImage -}}
{{ include "kfp-operator.trimmedContainerRegistry" . }}{{ .Values.manager.argo.compilerImage }}:0.6.0
{{- else -}}
{{ include "kfp-operator.trimmedContainerRegistry" . }}kfp-operator-argo-kfp-compiler:0.6.0
{{- end -}}
{{- end -}}

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
workflowTemplatePrefix: {{ include "kfp-operator.fullname" . }}-
{{- if .Values.manager.multiversion.enabled }}
multiversion: true
{{- end -}}
{{- end }}

{{- define "kfp-operator.configuration" -}}
{{ merge (include "configurationOverrides" . | fromYaml) .Values.manager.configuration (include "fallbackConfiguration" . | fromYaml) | toYaml }}
{{- end }}

{{- define "kfp-operator.argoNamespace" -}}
{{- (include "kfp-operator.configuration" . | fromYaml).workflowNamespace -}}
{{- end -}}

{{- define "kfp-operator.providerTypeExists" -}}
{{- range $providerName, $providerBlock := .Values.providers -}}
  {{- if eq $providerBlock.type $.ProviderType -}}{{ $providerName }}{{- end -}}
{{- end -}}
{{- end -}}
