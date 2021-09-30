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
Set namespace.
*/}}
{{- define "kfp-operator.namespace" -}}
{{- default (printf "%s-system" (include "kfp-operator.fullname" .)) .Values.namespace }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "kfp-operator.labels" -}}
control-plane: controller-manager
{{- end }}


{{/*
Create the name of the service account to use
*/}}
{{- define "kfp-operator.serviceAccountName" -}}
{{- default (printf "%s-controller-manager" (include "kfp-operator.fullname" .)) .Values.serviceAccountName }}
{{- end }}
