package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sky-uk/kfp-operator/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/server/resource"
)

func readinessHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Application is ready."))
}

func livenessHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Application is live."))
}

func createHandler(hr resource.HttpHandledResource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		resp, err := hr.Create(body)
		if err != nil {
			var userErr *resource.UserError
			if errors.As(err, &userErr) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusCreated)
		_, err = w.Write([]byte(resp.Id))
		if err != nil {
			http.Error(
				w,
				"Failed to write response body id",
				http.StatusInternalServerError,
			)
			return
		}
	}
}

func updateHandler(hr resource.HttpHandledResource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		decodedId, err := url.PathUnescape(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		resp, err := hr.Update(decodedId, body)
		if err != nil {
			var userErr *resource.UserError
			if errors.As(err, &userErr) {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		_, err = w.Write([]byte(resp.Id))
		if err != nil {
			http.Error(
				w,
				"Failed to write response body id",
				http.StatusInternalServerError,
			)
			return
		}
	}
}

func deleteHandler(a resource.HttpHandledResource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		decodedId, err := url.PathUnescape(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := a.Delete(decodedId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

func newHandler(resources []resource.HttpHandledResource) http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	mux.Get("/livez", livenessHandler)
	mux.Get("/readyz", readinessHandler)

	for _, resource := range resources {
		mux.Route("/resource/"+resource.Type(), func(r chi.Router) {
			r.Post("/", createHandler(resource))
			r.Put("/{id}", updateHandler(resource))
			r.Delete("/{id}", deleteHandler(resource))
		})
	}

	return mux
}

func Start(ctx context.Context, cfg config.Server, provider resource.Provider) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger := common.LoggerFromContext(ctx)

	httpResources := []resource.HttpHandledResource{
		&resource.Pipeline{Ctx: ctx, Provider: provider},
		&resource.Run{Ctx: ctx, Provider: provider},
		&resource.RunSchedule{Ctx: ctx, Provider: provider},
		&resource.Experiment{Ctx: ctx, Provider: provider},
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: newHandler(httpResources),
	}

	go func() {
		logger.Info(fmt.Sprintf("Starting HTTP server on %s", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err, "ListenAndServe", "failed")
		}
	}()

	return nil
}
