package provider

import (
	"regexp"
	"strings"

	"github.com/sky-uk/kfp-operator/provider-service/vai/internal/label"
)

type LabelSanitizer interface {
	Sanitize(labels map[string]string) map[string]string
}

type DefaultLabelSanitizer struct{}

// Vertex AI PipelineJob label keys and values can be no longer than 64 chars,
// can only contain lowercase letters, numeric chars, underscore and hyphens.
var vaiCompliant = regexp.MustCompilePOSIX(`[^a-z0-9_-]+`)

// Sanitize returns a processed copy of labels which are sanitized to be compliant
// to Vertex AI PipelineJob's labels restrictions.
func (dls DefaultLabelSanitizer) Sanitize(labels map[string]string) map[string]string {
	const maxLength = 63
	sanitized := make(map[string]string, len(labels))

	for kSan, vSan := range labels {
		kSan = strings.ToLower(kSan)
		vSan = strings.ToLower(vSan)

		// Differing methods of replacing invalid chars due to historical decisions.
		switch kSan {
		case label.PipelineVersion:
			vSan = vaiCompliant.ReplaceAllString(vSan, "-")
		case "schema_version", "sdk_version":
			vSan = vaiCompliant.ReplaceAllString(vSan, "_")
		default:
			kSan = vaiCompliant.ReplaceAllString(kSan, "")
			vSan = vaiCompliant.ReplaceAllString(vSan, "")
		}

		if len(kSan) > maxLength {
			kSan = kSan[:maxLength]
		}
		if len(vSan) > maxLength {
			vSan = vSan[:maxLength]
		}

		sanitized[kSan] = vSan
	}

	return sanitized
}
