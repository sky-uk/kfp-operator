package provider

import (
	"regexp"
	"strings"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/label"
)

type LabelSanitizer interface {
	Sanitize(labels map[string]string) map[string]string
}

type DefaultLabelSanitizer struct{}

// Vertex AI PipelineJob label keys and values can be no longer than 63 chars,
// can only contain lowercase letters, numeric chars, underscore and hyphens.
var vaiCompliant = regexp.MustCompilePOSIX(`[^a-z0-9_-]+`)

// Sanitize returns a processed copy of labels which are sanitized to be compliant
// to Vertex AI PipelineJob's labels restrictions.
func (dls DefaultLabelSanitizer) Sanitize(labels map[string]string) map[string]string {
	const maxLength = 63
	sanitized := make(map[string]string, len(labels))

	for k, v := range labels {
		key := strings.ToLower(k)
		value := strings.ToLower(v)

		// Differing methods of replacing invalid chars due to historical decisions.
		switch key {
		case label.PipelineVersion, label.SchemaVersion, label.SdkVersion:
			value = vaiCompliant.ReplaceAllString(value, "_")
		default:
			key = vaiCompliant.ReplaceAllString(key, "")
			value = vaiCompliant.ReplaceAllString(value, "")
		}

		if len(key) > maxLength {
			key = key[:maxLength]
		}
		if len(value) > maxLength {
			value = value[:maxLength]
		}

		sanitized[key] = value
	}

	return sanitized
}
