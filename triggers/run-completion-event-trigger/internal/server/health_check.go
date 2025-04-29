package server

type HealthCheck interface {
	IsHealthy() bool
}
