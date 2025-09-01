package provider

import (
	"encoding/json"
	"fmt"
	"maps"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"google.golang.org/protobuf/types/known/structpb"
)

type JobEnricher interface {
	Enrich(
		job *aiplatformpb.PipelineJob,
		raw resource.CompiledPipeline,
	) (*aiplatformpb.PipelineJob, error)
}

type DefaultJobEnricher struct {
	labelSanitizer LabelSanitizer
}

func NewDefaultJobEnricher() DefaultJobEnricher {
	return DefaultJobEnricher{
		labelSanitizer: DefaultLabelSanitizer{},
	}
}

func (dje DefaultJobEnricher) Enrich(
	job *aiplatformpb.PipelineJob,
	compiledPipeline resource.CompiledPipeline,
) (*aiplatformpb.PipelineJob, error) {
	var pipelineSpec map[string]any
	if err := json.Unmarshal(compiledPipeline.PipelineSpec, &pipelineSpec); err != nil {
		return nil, fmt.Errorf("failed to unmarshal PipelineSpec: %w", err)
	}

	pipelineSpecPb, err := structpb.NewStruct(pipelineSpec)
	if err != nil {
		return nil, err
	}

	if job.Labels == nil {
		job.Labels = map[string]string{}
	}

	schemaVersion, ok := pipelineSpec["schemaVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'schemaVersion', got %T",
			pipelineSpec["schemaVersion"],
		)
	}
	sdkVersion, ok := pipelineSpec["sdkVersion"].(string)
	if !ok {
		return nil, fmt.Errorf(
			"expected string for 'sdkVersion', got %T",
			pipelineSpec["sdkVersion"],
		)
	}

	job.Labels["schema_version"] = schemaVersion
	job.Labels["sdk_version"] = sdkVersion

	maps.Copy(job.Labels, compiledPipeline.Labels)

	job.Name = compiledPipeline.DisplayName
	job.Labels = dje.labelSanitizer.Sanitize(job.Labels)
	job.PipelineSpec = pipelineSpecPb
	return job, nil
}
