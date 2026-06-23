package provider

import (
	"maps"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"google.golang.org/protobuf/types/known/structpb"
)

type JobEnricher interface {
	Enrich(
		job *aiplatformpb.PipelineJob,
		raw map[string]any,
	) (*aiplatformpb.PipelineJob, error)
}

type DefaultJobEnricher struct {
	pipelineSchemaHandler PipelineSchemaHandler
	labelSanitizer        LabelSanitizer
}

func NewDefaultJobEnricher() DefaultJobEnricher {
	return DefaultJobEnricher{
		pipelineSchemaHandler: DefaultPipelineSchemaHandler{},
		labelSanitizer:        DefaultLabelSanitizer{},
	}
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

	job.Labels = dje.labelSanitizer.Sanitize(job.Labels)
	job.PipelineSpec = pv.pipelineSpec

	// Vertex binds RuntimeConfig per schemaVersion: Parameters for <= 2.0.0,
	// ParameterValues for >= 2.1.0. Submitting the wrong field returns
	// INTERNAL_ERROR.
	if pv.schemaVersion != nil && !pv.schemaVersion.LessThan(SchemaVersion2_1) {
		promoteParametersToValues(job)
	}
	return job, nil
}

// promoteParametersToValues copies RuntimeConfig.Parameters into
// RuntimeConfig.ParameterValues and clears Parameters.
func promoteParametersToValues(job *aiplatformpb.PipelineJob) {
	if job.RuntimeConfig == nil || len(job.RuntimeConfig.Parameters) == 0 {
		return
	}
	if job.RuntimeConfig.ParameterValues == nil {
		job.RuntimeConfig.ParameterValues = make(map[string]*structpb.Value, len(job.RuntimeConfig.Parameters))
	}
	for name, value := range job.RuntimeConfig.Parameters {
		switch v := value.GetValue().(type) {
		case *aiplatformpb.Value_StringValue:
			job.RuntimeConfig.ParameterValues[name] = structpb.NewStringValue(v.StringValue)
		case *aiplatformpb.Value_IntValue:
			job.RuntimeConfig.ParameterValues[name] = structpb.NewNumberValue(float64(v.IntValue))
		case *aiplatformpb.Value_DoubleValue:
			job.RuntimeConfig.ParameterValues[name] = structpb.NewNumberValue(v.DoubleValue)
		default:
			job.RuntimeConfig.ParameterValues[name] = structpb.NewNullValue()
		}
	}
	job.RuntimeConfig.Parameters = nil
}
