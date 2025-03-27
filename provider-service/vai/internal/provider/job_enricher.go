package provider

import (
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
	for k, v := range pv.labels {
		job.Labels[k] = v
	}
	job.PipelineSpec = pv.pipelineSpec
	return job, nil
}
