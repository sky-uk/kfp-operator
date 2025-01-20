package client

import (
	"context"

	"cloud.google.com/go/aiplatform/apiv1/aiplatformpb"
	"github.com/googleapis/gax-go/v2"
)

type PipelineJobClient interface {
	GetPipelineJob(
		ctx context.Context,
		req *aiplatformpb.GetPipelineJobRequest,
		opts ...gax.CallOption,
	) (*aiplatformpb.PipelineJob, error)

	CreatePipelineJob(
		ctx context.Context,
		req *aiplatformpb.CreatePipelineJobRequest,
		opts ...gax.CallOption,
	) (*aiplatformpb.PipelineJob, error)
}
