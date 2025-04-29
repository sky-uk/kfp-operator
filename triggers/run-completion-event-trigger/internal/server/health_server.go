package server

import (
	"context"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type HealthServer struct {
	healthpb.UnimplementedHealthServer
	HealthCheck HealthCheck
}

func (hs *HealthServer) Check(
	ctx context.Context,
	_ *healthpb.HealthCheckRequest,
) (*healthpb.HealthCheckResponse, error) {
	if hs.HealthCheck.IsHealthy() {
		return &healthpb.HealthCheckResponse{
			Status: healthpb.HealthCheckResponse_SERVING,
		}, nil
	}
	return &healthpb.HealthCheckResponse{
		Status: healthpb.HealthCheckResponse_NOT_SERVING,
	}, nil
}
