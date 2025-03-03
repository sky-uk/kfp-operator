package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sky-uk/kfp-operator/argo/common"
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
			writeErrorResponse(w, errors.New("failed to read request body"), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		resp, err := hr.Create(body)
		if err != nil {
			var userErr *resource.UserError
			if errors.As(err, &userErr) {
				writeErrorResponse(w, err, http.StatusBadRequest)
				return
			} else {
				writeErrorResponse(w, err, http.StatusInternalServerError)
				return
			}
		}

		writeResponse(w, resp, http.StatusCreated)
	}
}

func updateHandler(hr resource.HttpHandledResource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		decodedId, err := url.PathUnescape(id)
		if err != nil {
			writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrorResponse(w, errors.New("failed to read request body"), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		resp, err := hr.Update(decodedId, body)
		if err != nil {
			var userErr *resource.UserError
			if errors.As(err, &userErr) {
				writeErrorResponse(w, err, http.StatusBadRequest)
				return
			} else {
				writeErrorResponse(w, err, http.StatusInternalServerError)
				return
			}
		}

		writeResponse(w, resp, http.StatusOK)
	}
}

func deleteHandler(a resource.HttpHandledResource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		decodedId, err := url.PathUnescape(id)
		if err != nil {
			writeErrorResponse(w, err, http.StatusBadRequest)
			return
		}

		err = a.Delete(decodedId)
		if err != nil {
			writeErrorResponse(w, err, http.StatusInternalServerError)
			return
		}

		writeResponse(w, resource.ResponseBody{Id: decodedId}, http.StatusOK)
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

func writeErrorResponse(w http.ResponseWriter, providerError error, statusCode int) {
	responseBody := resource.ResponseBody{
		ProviderError: providerError.Error(),
	}
	writeResponse(w, responseBody, statusCode)
}

func writeResponse(w http.ResponseWriter, responseBody resource.ResponseBody, statusCode int) {
	marshalledResponse, err := json.Marshal(responseBody)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(statusCode)
	if _, err = w.Write(marshalledResponse); err != nil {
		http.Error(w, "failed to write request body", http.StatusInternalServerError)
		return
	}
}
