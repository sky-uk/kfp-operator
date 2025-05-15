package server

import "context"

type HealthCheck interface {
	Name() string
	IsHealthy(ctx context.Context) bool
}
