package main

import (
	"context"
	"github.com/kubeflow/pipelines/backend/api/go_client"
)

type KfpApi interface {
	GetRunConfiguration(ctx context.Context, runId string) (string, error)
}

type GrpcKfpApi struct {
	RunServiceClient go_client.RunServiceClient
}

func (gka *GrpcKfpApi) GetRunConfiguration(ctx context.Context, runId string) (string, error) {
	runDetail, err := gka.RunServiceClient.GetRun(ctx, &go_client.GetRunRequest{RunId: runId})
	if err != nil {
		return "", err
	}

	var runConfigName string
	for _, ref := range runDetail.GetRun().GetResourceReferences() {
		if ref.GetKey().GetType() == go_client.ResourceType_JOB && ref.GetRelationship() == go_client.Relationship_CREATOR {
			runConfigName = ref.GetName()
			break
		}
	}

	return runConfigName, nil
}
