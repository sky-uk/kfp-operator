package server

import (
	"context"

	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

func Start(ctx context.Context, cfg config.Config, provider resource.Provider, healthChecks []HealthCheck) error {
	err := MetricsServer{}.Start(ctx, cfg.Metrics, cfg.ProviderName)
	if err != nil {
		return err
	}

	err = ProviderServer{}.Start(ctx, cfg.Server, provider, healthChecks)
	if err != nil {
		return err
	}
	return nil
}
