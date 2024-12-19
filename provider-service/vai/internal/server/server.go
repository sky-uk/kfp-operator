package server

import (
	"context"
	"fmt"
	"net/http"
	"github.com/sky-uk/kfp-operator/argo/common"
	"github.com/sky-uk/kfp-operator/provider-service/base/pkg/config"
)

func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func Start(
	ctx context.Context,
	config config.Server,
) error {
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", config.Host, config.Port),
		Handler: http.HandlerFunc(handler),
	}

	l := common.LoggerFromContext(ctx)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		l.Error(err, "HTTP server failed")
	}

	return nil
}
