package server

import (
	"context"
	"fmt"

	"github.com/sky-uk/kfp-operator/argo/common"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer struct {
	healthpb.UnimplementedHealthServer
	HealthCheck HealthCheck
}

func (hs *HealthServer) Check(
	ctx context.Context,
	req *healthpb.HealthCheckRequest,
) (*healthpb.HealthCheckResponse, error) {
	log := common.LoggerFromContext(ctx)

	switch req.GetService() {
	case "liveness":
		return &healthpb.HealthCheckResponse{
			Status: healthpb.HealthCheckResponse_SERVING,
		}, nil
	case "readiness":
		if hs.HealthCheck.IsHealthy() {
			return &healthpb.HealthCheckResponse{
				Status: healthpb.HealthCheckResponse_SERVING,
			}, nil
		}
		log.Error(
			fmt.Errorf("service is not ready"),
			"dependency",
			"name",
			hs.HealthCheck.Name(),
		)

		return &healthpb.HealthCheckResponse{
			Status: healthpb.HealthCheckResponse_NOT_SERVING,
		}, nil
	default:
		return &healthpb.HealthCheckResponse{
				Status: healthpb.HealthCheckResponse_SERVICE_UNKNOWN,
			}, fmt.Errorf(
				"Unexpected service name [%s]. Expected 'liveness' or 'readiness'",
				req.GetService(),
			)
	}
}
