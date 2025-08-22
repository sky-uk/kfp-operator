package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sky-uk/kfp-operator/argo/providers/base"
)

// CompiledPipeline represents the output of the python compile step, and
// describes what vertex ai or kubeflow pipelines should do.
type PipelineDefinitionWrapper struct {
	PipelineDefinition base.PipelineDefinition `json:"pipelineDefinition"`
	CompiledPipeline   json.RawMessage         `json:"compiledPipeline,omitempty"`
}

type Provider interface {
	PipelineProvider
	RunProvider
	RunScheduleProvider
	ExperimentProvider
}

type PipelineProvider interface {
	CreatePipeline(ctx context.Context, ppd PipelineDefinitionWrapper) (string, error)
	UpdatePipeline(ctx context.Context, ppd PipelineDefinitionWrapper, id string) (string, error)
	DeletePipeline(ctx context.Context, id string) error
}

type RunProvider interface {
	CreateRun(ctx context.Context, rd base.RunDefinition) (string, error)
	DeleteRun(ctx context.Context, id string) error
}

type RunScheduleProvider interface {
	CreateRunSchedule(ctx context.Context, rsd base.RunScheduleDefinition) (string, error)
	UpdateRunSchedule(ctx context.Context, rsd base.RunScheduleDefinition, id string) (string, error)
	DeleteRunSchedule(ctx context.Context, id string) error
}

type ExperimentProvider interface {
	CreateExperiment(ctx context.Context, ed base.ExperimentDefinition) (string, error)
	UpdateExperiment(ctx context.Context, ed base.ExperimentDefinition, id string) (string, error)
	DeleteExperiment(ctx context.Context, id string) error
}

type UserError struct {
	E error
}

func (e *UserError) Error() string {
	return e.E.Error()
}

type UnimplementedError struct {
	Method       string
	ResourceType string
}

func (e *UnimplementedError) Error() string {
	return fmt.Sprintf("Method %s unimplemented for resource %s", e.Method, e.ResourceType)
}
