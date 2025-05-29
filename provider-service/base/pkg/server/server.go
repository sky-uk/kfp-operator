package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"golang.org/x/sync/errgroup"
)

func Start(ctx context.Context, cfg config.Config, provider resource.Provider) error {
	log := common.LoggerFromContext(ctx)

	meterProvider, err := initMeterProvider(cfg.ProviderName)
	if err != nil {
		return err
	}

	providerServer := NewProviderServer(ctx, cfg.Server, provider)
	metricsServer := NewMetricsServer(cfg.Server.Host, cfg.Metrics.Port)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(
		func() error {
			addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
			log.Info("Provider Server starting", "addr", addr)

			err := providerServer.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			log.Error(err, "Provider Server failed to start or crashed")
			return err
		},
	)

	g.Go(
		func() error {
			addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Metrics.Port)
			log.Info("Metrics Server starting", "addr", addr)

			err := metricsServer.ListenAndServe()
			if errors.Is(err, http.ErrServerClosed) {
				return nil
			}

			log.Error(err, "Metrics Server failed to start or crashed")
			return err
		},
	)

	g.Go(
		func() error {
			<-ctx.Done()

			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			log.Info("Shutting down servers...")

			return errors.Join(
				providerServer.Shutdown(shutdownCtx),
				metricsServer.Shutdown(shutdownCtx),
				meterProvider.Shutdown(shutdownCtx),
			)
		},
	)

	if err := g.Wait(); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	log.Info("Servers shutdown cleanly")

	return nil
}
