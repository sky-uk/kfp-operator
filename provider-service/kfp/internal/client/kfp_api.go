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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

// dialer is the subset of grpc.NewClient's signature used to open the
// connection, so the real dialer can be swapped out (for example, in tests).
type dialer func(
	target string,
	opts ...grpc.DialOption,
) (*grpc.ClientConn, error)

// bearerTokenPath is the fixed path at which the KFP provider-service reads the
// projected ServiceAccount token when running in multi-user mode. Provider
// owners must mount their projected token at this path via the Provider CR's
// podTemplateVolumes/podTemplateVolumeMounts.
const bearerTokenPath = "/var/run/secrets/kfp/token"

// MultiUserTokenSource returns the bearer-token source for cfg and true when
// multi-user mode is enabled, reading the projected token from the fixed
// bearerTokenPath. It returns (nil, false) when multi-user mode is disabled.
func MultiUserTokenSource(cfg config.Config) (auth.TokenSource, bool) {
	if !cfg.Parameters.KfpMultiUserMode {
		return nil, false
	}
	return auth.NewFileTokenSource(bearerTokenPath), true
}

// GrpcDialOptions returns the gRPC dial options for connecting to the KFP API
// with cfg: insecure transport credentials always, plus a bearer-token
// credential read from the projected token when multi-user mode is enabled.
func GrpcDialOptions(cfg config.Config) []grpc.DialOption {
	dialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if tokenSource, ok := MultiUserTokenSource(cfg); ok {
		dialOptions = append(dialOptions, auth.BearerDialOption(tokenSource))
	}
	return dialOptions
}

func CreateKfpApi(ctx context.Context, cfg config.Config) (*GrpcKfpApi, error) {
	if tokenSource, ok := MultiUserTokenSource(cfg); ok {
		if _, err := tokenSource.Token(); err != nil {
			return nil, fmt.Errorf("multi-user mode requires a readable bearer token: %w", err)
		}
	}
	return createKfpApi(ctx, cfg, grpc.NewClient)
}

func createKfpApi(
	ctx context.Context,
	cfg config.Config,
	dial dialer,
) (*GrpcKfpApi, error) {
	logger := logr.FromContextOrDiscard(ctx)

	kfpApi, err := connectToKfpApi(
		dial,
		cfg.Parameters.GrpcKfpApiAddress,
		GrpcDialOptions(cfg)...,
	)
	if err != nil {
		logger.Error(err, "failed to connect to Kubeflow API", "address", cfg.Parameters.GrpcKfpApiAddress)
		return nil, err
	}
	return kfpApi, nil
}

func connectToKfpApi(
	dial dialer,
	address string,
	dialOptions ...grpc.DialOption,
) (*GrpcKfpApi, error) {
	conn, err := dial(address, dialOptions...)
	if err != nil {
		return nil, err
	}

	return &GrpcKfpApi{
		RunServiceClient: go_client.NewRunServiceClient(conn),
	}, nil
}
