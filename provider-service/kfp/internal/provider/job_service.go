package provider

import (
	"context"

	"github.com/kubeflow/pipelines/backend/api/go_client"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type JobService interface {
	DeleteJob(id string) error
}

type DefaultJobService struct {
	ctx    context.Context
	client client.JobServiceClient
}

func (js *DefaultJobService) DeleteJob(id string) error {
	_, err := js.client.DeleteJob(
		js.ctx,
		&go_client.DeleteJobRequest{
			Id: id,
		},
	)
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			// is a gRPC error
			switch st.Code() {
			case codes.NotFound:
				return nil
			default:
				return err
			}
		}
		return err
	}

	return nil
}
