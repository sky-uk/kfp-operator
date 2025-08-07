package provider

import (
	"maps"
	"regexp"
	"strings"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
)

type JobEnricher interface {
	Enrich(
		job *aiplatformpb.PipelineJob,
		raw map[string]any,
	) (*aiplatformpb.PipelineJob, error)
}

type DefaultJobEnricher struct {
	pipelineSchemaHandler PipelineSchemaHandler
}

func (dje DefaultJobEnricher) Enrich(
	job *aiplatformpb.PipelineJob,
	raw map[string]any,
) (*aiplatformpb.PipelineJob, error) {
	pv, err := dje.pipelineSchemaHandler.extract(raw)
	if err != nil {
		return nil, err
	}
	job.Name = pv.name
	if job.Labels == nil {
		job.Labels = map[string]string{}
	}

	maps.Copy(job.Labels, pv.labels)

	job.Labels = sanitizeLabels(job.Labels)
	job.PipelineSpec = pv.pipelineSpec
	return job, nil
}

// Mutates PipelineJob labels in-place
// Combines the `raw` labels with the existing job labels as well; both are
// sanitized
var regex = regexp.MustCompile(`[^a-z0-9_-]+`)

func sanitizeLabels(labels map[string]string) map[string]string {

	const maxLength = 63
	sanitized := make(map[string]string, len(labels))

	for kSan, vSan := range labels {
		kSan = strings.ToLower(kSan)
		vSan = strings.ToLower(vSan)

		switch kSan {
		case "schema_version", "sdk_version":
			vSan = regex.ReplaceAllString(vSan, "_")
		default:
			kSan = regex.ReplaceAllString(kSan, "")
			vSan = regex.ReplaceAllString(vSan, "")
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
