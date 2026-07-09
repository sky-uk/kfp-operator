{{/*
Namespace the provider workflow resources are created in. Defaults to the
release namespace when .Values.namespace is empty.
*/}}
{{- define "kfp-provider-workflows.namespace" -}}
{{- default .Release.Namespace .Values.namespace -}}
{{- end -}}

{{/*
Name of the Provider resource. Defaults to the release name.
*/}}
{{- define "kfp-provider-workflows.providerName" -}}
{{- default .Release.Name .Values.provider.name -}}
{{- end -}}

{{/*
Only render object if not empty.
*/}}
{{- define "kfp-provider-workflows.notEmptyYaml" -}}
{{- if . }}
{{ . | toYaml }}
{{- end }}
{{- end }}
