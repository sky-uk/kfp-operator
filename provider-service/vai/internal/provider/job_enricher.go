package provider

import (
	"fmt"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/protobuf/types/known/structpb"
)

type JobEnricher interface {
	Enrich(
		job *aiplatformpb.PipelineJob,
		raw map[string]any,
	) (*aiplatformpb.PipelineJob, error)
}

type DefaultJobEnricher struct{}

// Enrich enriches a pipeline job with specific data from the compiled pipeline;
// which is represented by an unstructured map. However it expects the values to
// have a particular structure.
func (je DefaultJobEnricher) Enrich(
	job *aiplatformpb.PipelineJob,
	raw map[string]any,
) (*aiplatformpb.PipelineJob, error) {
	displayName, ok := raw["displayName"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'displayName', got %T",
			raw["displayName"],
		)
	}
	job.DisplayName = displayName

	pipelineSpec, ok := raw["pipelineSpec"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"expected map for 'pipelineSpec', got %T",
			raw["pipelineSpec"],
		)
	}

	pipelineSpecStruct, err := structpb.NewStruct(pipelineSpec)
	if err != nil {
		return nil, err
	}
	job.PipelineSpec = pipelineSpecStruct

	labels, ok := raw["labels"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"expected map for 'labels', got %T",
			raw["labels"],
		)
	}
	if job.Labels == nil {
		job.Labels = map[string]string{}
	}
	for k, v := range labels {
		job.Labels[k] = v.(string)
	}

	runtimeConfig, ok := raw["runtimeConfig"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf(
			"expected map for 'runtimeConfig', got %T",
			raw["runtimeConfig"],
		)
	}

	gcsOutputDirectory, ok := runtimeConfig["gcsOutputDirectory"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'gcsOutputDirectory', got %T",
			runtimeConfig["gcsOutputDirectory"],
		)
	}

	if job.RuntimeConfig == nil {
		return nil, fmt.Errorf("job has no RuntimeConfig")
	}
	job.RuntimeConfig.GcsOutputDirectory = gcsOutputDirectory

	return job, nil
}
