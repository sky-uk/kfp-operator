package label

import (
	"fmt"
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
// Trim the values to a maximum length of 64 characters when combined with keys
// Enforce lowercase on all key, values
// Remove special chars
func SanitizeLabels(labels map[string]string) map[string]string {
	newLabels := map[string]string{}
	regex := regexp.MustCompile(`[^a-z0-9_-]+`)
	for k, v := range labels {
		k = strings.ToLower(k)
		v = strings.ToLower(v)

		k = regex.ReplaceAllString(k, "")
		v = regex.ReplaceAllString(v, "")

		maxLength := 64
		if len(fmt.Sprintf("%s: %s", k, v)) >= maxLength {
			keyLength := len(fmt.Sprintf("%s: ", k))
			elipsesLength := 3
			trimLength := maxLength - keyLength - elipsesLength
			v = v[:trimLength] + "..."
		}

		newLabels[k] = v
	}

	return newLabels
}
