package server

import (
	"context"
	"fmt"

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
