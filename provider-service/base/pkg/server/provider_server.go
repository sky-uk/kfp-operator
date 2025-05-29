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
	"github.com/go-logr/logr"
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

func createHandler(ctx context.Context, hr resource.HttpHandledResource) http.HandlerFunc {
	logger := common.LoggerFromContext(ctx)

	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := logr.NewContext(r.Context(), logger)

		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrorResponse(w, "", errors.New("failed to read request body"), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		resp, err := hr.Create(requestCtx, body)

		switch {
		case err == nil:
			writeResponse(w, resp, http.StatusCreated)
			return

		case errors.As(err, new(*resource.UserError)):
			writeErrorResponse(w, "", err, http.StatusBadRequest)
			return

		case errors.As(err, new(*resource.UnimplementedError)):
			writeErrorResponse(w, "", err, http.StatusNotImplemented)
			return

		default:
			writeErrorResponse(w, "", err, http.StatusInternalServerError)
			return
		}
	}
}

func updateHandler(ctx context.Context, hr resource.HttpHandledResource) http.HandlerFunc {
	logger := common.LoggerFromContext(ctx)

	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := logr.NewContext(r.Context(), logger)

		id := chi.URLParam(r, "id")
		decodedId, err := url.PathUnescape(id)
		if err != nil {
			writeErrorResponse(w, decodedId, err, http.StatusBadRequest)
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeErrorResponse(w, decodedId, errors.New("failed to read request body"), http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		resp, err := hr.Update(requestCtx, decodedId, body)

		switch {
		case err == nil:
			writeResponse(w, resp, http.StatusOK)
			return

		case errors.As(err, new(*resource.UserError)):
			writeErrorResponse(w, decodedId, err, http.StatusBadRequest)
			return

		case errors.As(err, new(*resource.UnimplementedError)):
			writeErrorResponse(w, decodedId, err, http.StatusNotImplemented)
			return

		default:
			writeErrorResponse(w, decodedId, err, http.StatusInternalServerError)
			return
		}
	}
}

func deleteHandler(ctx context.Context, a resource.HttpHandledResource) http.HandlerFunc {
	logger := common.LoggerFromContext(ctx)

	return func(w http.ResponseWriter, r *http.Request) {
		requestCtx := logr.NewContext(r.Context(), logger)

		id := chi.URLParam(r, "id")
		decodedId, err := url.PathUnescape(id)
		if err != nil {
			writeErrorResponse(w, decodedId, err, http.StatusBadRequest)
			return
		}

		err = a.Delete(requestCtx, decodedId)

		switch {
		case err == nil:
			writeResponse(w, resource.ResponseBody{}, http.StatusOK)
			return

		case errors.As(err, new(*resource.UserError)):
			writeErrorResponse(w, decodedId, err, http.StatusBadRequest)
			return

		case errors.As(err, new(*resource.UnimplementedError)):
			writeErrorResponse(w, decodedId, err, http.StatusNotImplemented)
			return

		default:
			writeErrorResponse(w, decodedId, err, http.StatusInternalServerError)
			return
		}
	}
}

func newHandler(ctx context.Context, resources []resource.HttpHandledResource) http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)

	mux.Get("/livez", livenessHandler)
	mux.Get("/readyz", readinessHandler)

	for _, resource := range resources {
		mux.Route("/resource/"+resource.Type(), func(r chi.Router) {
			r.Post("/", createHandler(ctx, resource))
			r.Put("/{id}", updateHandler(ctx, resource))
			r.Delete("/{id}", deleteHandler(ctx, resource))
		})
	}

	return mux
}

func NewProviderServer(
	ctx context.Context,
	cfg config.Server,
	provider resource.Provider,
) *http.Server {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	httpResources := []resource.HttpHandledResource{
		&resource.Pipeline{Provider: provider},
		&resource.Run{Provider: provider},
		&resource.RunSchedule{Provider: provider},
		&resource.Experiment{Provider: provider},
	}

	return &http.Server{
		Addr:    addr,
		Handler: newHandler(ctx, httpResources),
	}
}

type ProviderServer struct{}

func (ps ProviderServer) Start(ctx context.Context, cfg config.Server, provider resource.Provider) error {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	logger := common.LoggerFromContext(ctx)

	httpResources := []resource.HttpHandledResource{
		&resource.Pipeline{Provider: provider},
		&resource.Run{Provider: provider},
		&resource.RunSchedule{Provider: provider},
		&resource.Experiment{Provider: provider},
	}

	srv := &http.Server{
		Addr:    addr,
		Handler: newHandler(ctx, httpResources),
	}

	go func() {
		logger.Info(fmt.Sprintf("Starting HTTP server on %s", addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err, "ListenAndServe", "failed")
		}
	}()

	return nil
}

func writeErrorResponse(w http.ResponseWriter, id string, providerError error, statusCode int) {
	responseBody := resource.ResponseBody{
		Id:            id,
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
	w.Write(marshalledResponse)
	return

}
