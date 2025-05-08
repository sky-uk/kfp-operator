package server

type HealthCheck interface {
	Name() string
	IsHealthy() bool
}
