package server

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	chi "github.com/go-chi/chi/v5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
)

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Application is ready."))
}

func livenessHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Application is live."))
}

func New() http.Handler {
	mux := chi.NewRouter()
	mux.Get("/livez", livenessHandler)
	mux.Get("/readyz", readinessHandler)

	return mux
}

// Start starts the HTTP server and gracefully handles shutdown
func Start(ctx context.Context, cfg config.Server) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger := common.LoggerFromContext(ctx)
	// Create a server with the provided address and handler
	srv := &http.Server{
		Addr:    addr,
		Handler: New(),
	}

	// Start the server in a goroutine
	go func() {
		logger.Info("Starting HTTP server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error(err, "ListenAndServe", "failed")
		}
	}()

	// Wait for termination signal (Ctrl+C, kill, etc.)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Wait for a signal to gracefully shutdown the server
	<-sigCh
	logger.Info("Shutting down server...")

	// Graceful shutdown with a timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error(err, "shutdown", "forced")
	}
	logger.Info("Server gracefully stopped")

	return nil
}
