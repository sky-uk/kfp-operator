package label

import (
	"regexp"
	"strings"
)

const (
	ProviderName              = "provider-name"
	ProviderNamespace         = "provider-namespace"
	PipelineName              = "pipeline-name"
	PipelineNamespace         = "pipeline-namespace"
	PipelineVersion           = "pipeline-version"
	RunConfigurationName      = "runconfiguration-name"
	RunConfigurationNamespace = "runconfiguration-namespace"
	RunName                   = "run-name"
	RunNamespace              = "run-namespace"
)

// SanitizeLabels takes a map of labels and sanitizes them according to the following rules:
// Trim the values and keys to a maximum length of 63 characters
// Enforce lowercase on all key, values
// Remove special chars
func SanitizeLabels(labels map[string]string) map[string]string {
	newLabels := map[string]string{}
	regex := regexp.MustCompile(`[^a-z0-9_-]+`)
	maxLength := 63
	for k, v := range labels {
		k = strings.ToLower(k)
		v = strings.ToLower(v)

		k = regex.ReplaceAllString(k, "")
		v = regex.ReplaceAllString(v, "")

		if len(k) > maxLength {
			k = k[:maxLength]
		}
		if len(v) > maxLength {
			v = v[:maxLength]
		}
		newLabels[k] = v
	}

	return newLabels
}
