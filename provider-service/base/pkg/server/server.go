package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
	"io"
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

func postHandler(a resource.HttpHandledResource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		resp, err := a.Create(body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(resp.Id))
	}
}

func deleteHandler(a resource.HttpHandledResource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		err := a.Delete(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func putHandler(a resource.HttpHandledResource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		err = a.Update(id, body)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func New(resources []resource.HttpHandledResource) http.Handler {
	mux := chi.NewRouter()
	mux.Get("/livez", livenessHandler)
	mux.Get("/readyz", readinessHandler)

	for _, resource := range resources {
		mux.Route("/resource/"+resource.Name(), func(r chi.Router) {
			r.Post("/", postHandler(resource))
			r.Put("/{id}", putHandler(resource))
			r.Delete("/{id}", deleteHandler(resource))
		})
	}

	return mux
}

// Start starts the HTTP server and gracefully handles shutdown
func Start(ctx context.Context, cfg config.Server, provider resource.Provider) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger := common.LoggerFromContext(ctx)

	httpResources := []resource.HttpHandledResource{
		&resource.Pipeline{Provider: provider},
		&resource.Run{Provider: provider},
		&resource.RunSchedule{Provider: provider},
		&resource.Experiment{Provider: provider},
	}
	// Create a server with the provided address and handler
	srv := &http.Server{
		Addr:    addr,
		Handler: New(httpResources),
	}

	// Start the server in a goroutine
	go func() {
		logger.Info("Starting HTTP server on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
