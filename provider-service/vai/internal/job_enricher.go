package internal

import (
	"fmt"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/protobuf/types/known/structpb"
)

type JobEnricher struct{}

func (je JobEnricher) enrich(
	job *aiplatformpb.PipelineJob,
	raw map[string]any,
) error {
	displayName, ok := raw["displayName"].(string)
	if !ok {
		return fmt.Errorf("expected string for 'displayName', got %T", raw["displayName"])
	}
	job.DisplayName = displayName

	pipelineSpec, ok := raw["pipelineSpec"].(map[string]any)
	if !ok {
		return fmt.Errorf("expected map for 'pipelineSpec', got %T", raw["pipelineSpec"])
	}

	pipelineSpecStruct, err := structpb.NewStruct(pipelineSpec)
	if err != nil {
		return err
	}
	job.PipelineSpec = pipelineSpecStruct

	labels, ok := raw["labels"].(map[string]any)
	if !ok {
		return fmt.Errorf("expected map for 'labels', got %T", raw["labels"])
	}
	if job.Labels == nil {
		job.Labels = map[string]string{}
	}
	for k, v := range labels {
		job.Labels[k] = v.(string)
	}

	runtimeConfig, ok := raw["runtimeConfig"].(map[string]any)
	if !ok {
		return fmt.Errorf("expected map for 'runtimeConfig', got %T", raw["runtimeConfig"])
	}

	gcsOutputDirectory, ok := runtimeConfig["gcsOutputDirectory"].(string)
	if !ok {
		return fmt.Errorf(
			"expected string for 'gcsOutputDirectory', got %T",
			runtimeConfig["gcsOutputDirectory"],
		)
	}

	if job.RuntimeConfig == nil {
		return fmt.Errorf("job has no RuntimeConfig")
	}
	job.RuntimeConfig.GcsOutputDirectory = gcsOutputDirectory
	return nil
}
