package provider

import (
	"maps"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

type JobEnricher interface {
	Enrich(
		job *aiplatformpb.PipelineJob,
		raw resource.CompiledPipeline,
	) (*aiplatformpb.PipelineJob, error)
}

type DefaultJobEnricher struct {
	pipelineSchemaHandler PipelineSchemaHandler
	labelSanitizer        LabelSanitizer
}

func NewDefaultJobEnricher() DefaultJobEnricher {
	return DefaultJobEnricher{
		pipelineSchemaHandler: SchemaHandler{},
		labelSanitizer:        DefaultLabelSanitizer{},
	}
}

func (dje DefaultJobEnricher) Enrich(
	job *aiplatformpb.PipelineJob,
	compiledPipeline resource.CompiledPipeline,
) (*aiplatformpb.PipelineJob, error) {
	pv, err := dje.pipelineSchemaHandler.extract(compiledPipeline)
	if err != nil {
		return nil, err
	}
	job.Name = pv.name
	if job.Labels == nil {
		job.Labels = map[string]string{}
	}

	maps.Copy(job.Labels, pv.labels)

	job.Labels = dje.labelSanitizer.Sanitize(job.Labels)
	job.PipelineSpec = pv.pipelineSpec
	return job, nil
}
