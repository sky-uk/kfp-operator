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
Name of the provider service ServiceAccount. Defaults to
kfp-provider-<provider-name>.
*/}}
{{- define "kfp-provider-workflows.providerServiceAccountName" -}}
{{- default (printf "kfp-provider-%s" (include "kfp-provider-workflows.providerName" .)) .Values.provider.serviceAccount.name -}}
{{- end -}}

{{/*
Only render object if not empty.
*/}}
{{- define "kfp-provider-workflows.notEmptyYaml" -}}
{{- if . }}
{{ . | toYaml }}
{{- end }}
{{- end }}
