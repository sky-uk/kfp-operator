package client

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/auth"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/label"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/client/resource"
	"github.com/sky-uk/kfp-operator/provider-service/kfp/internal/config"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/transport"

	"github.com/kubeflow/pipelines/backend/api/v2beta1/go_client"
)

type GrpcKfpApi struct {
	RunServiceClient
}

func (gka *GrpcKfpApi) GetResourceReferences(ctx context.Context, runId string) (resource.References, error) {
	resourceReferences := resource.References{}

	run, err := gka.RunServiceClient.GetRun(ctx, &go_client.GetRunRequest{RunId: runId})
	if err != nil {
		return resource.References{}, err
	}

	params := run.RuntimeConfig.GetParameters()
	if params != nil {
		if runName, ok := params[label.RunName]; ok {
			resourceReferences.RunName.Name = runName.GetStringValue()
		}

		if runNamespace, ok := params[label.RunNamespace]; ok {
			resourceReferences.RunName.Namespace = runNamespace.GetStringValue()
		}

		if runConfigurationName, ok := params[label.RunConfigurationName]; ok {
			resourceReferences.RunConfigurationName.Name = runConfigurationName.GetStringValue()
		}

		if runConfigurationNamespace, ok := params[label.RunConfigurationNamespace]; ok {
			resourceReferences.RunConfigurationName.Namespace = runConfigurationNamespace.GetStringValue()
		}

		if pipelineName, ok := params[label.PipelineName]; ok {
			resourceReferences.PipelineName.Name = pipelineName.GetStringValue()
		}

		if pipelineNamespace, ok := params[label.PipelineNamespace]; ok {
			resourceReferences.PipelineName.Namespace = pipelineNamespace.GetStringValue()
		}
	}

	if run.CreatedAt != nil {
		runCreateTime := run.CreatedAt.AsTime()
		resourceReferences.CreatedAt = &runCreateTime
	}

	if run.FinishedAt != nil {
		runFinishedTime := run.FinishedAt.AsTime()
		resourceReferences.FinishedAt = &runFinishedTime
	} else {
		// the FinishededAt field is unreliable, so if it's nil we set it to the current time
		currentTime := time.Now().UTC()
		resourceReferences.FinishedAt = &currentTime
	}
	return resourceReferences, nil
}

func CreateKfpApi(ctx context.Context, cfg config.Config) (*GrpcKfpApi, error) {
	logger := logr.FromContextOrDiscard(ctx)

	tokenSource, err := MultiUserTokenSource(cfg)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.NewClient(
		cfg.Parameters.GrpcKfpApiAddress,
		GrpcDialOptions(tokenSource)...,
	)
	if err != nil {
		logger.Error(err, "failed to connect to Kubeflow API", "address", cfg.Parameters.GrpcKfpApiAddress)
		return nil, err
	}

	return &GrpcKfpApi{
		RunServiceClient: go_client.NewRunServiceClient(conn),
	}, nil
}

// bearerTokenPath is where the projected ServiceAccount token must be mounted
// in multi-user mode.
const bearerTokenPath = "/var/run/secrets/kfp/token"

// MultiUserTokenSource returns a validated, caching bearer-token source in
// multi-user mode, or a nil source when it is disabled.
func MultiUserTokenSource(cfg config.Config) (oauth2.TokenSource, error) {
	if !cfg.Parameters.KfpMultiUserMode {
		return nil, nil
	}
	tokenSource := transport.NewCachedFileTokenSource(bearerTokenPath)
	if _, err := tokenSource.Token(); err != nil {
		return nil, fmt.Errorf("multi-user mode requires a readable bearer token: %w", err)
	}
	return tokenSource, nil
}

// GrpcDialOptions returns the KFP gRPC dial options, adding a bearer
// credential when tokenSource is non-nil.
func GrpcDialOptions(tokenSource oauth2.TokenSource) []grpc.DialOption {
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if tokenSource != nil {
		dialOptions = append(dialOptions, auth.BearerDialOption(tokenSource))
	}
	return dialOptions
}
