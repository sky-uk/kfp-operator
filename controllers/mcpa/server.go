package mcpa

import (
	"context"
	"encoding/json"
	"net/http"

	v1beta1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/go-logr/logr"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

//
// Runnable (UNCHANGED)
//

type Runnable struct {
	Server *MCPServer
}

func (r *Runnable) Start(ctx context.Context) error {
	go func() {
		logr.FromContextOrDiscard(ctx).Info("MCP Server started")
		if err := r.Server.Start(); err != nil {
			panic(err)
		}
	}()
	<-ctx.Done()
	return nil
}

//
// MCPServer
//

type MCPServer struct {
	Cache cache.Cache
}

func NewMCPServer(c cache.Cache) *MCPServer {
	return &MCPServer{Cache: c}
}

//
// Start MCP server (SDK-managed MCP over HTTP)
//

func (s *MCPServer) Start() error {
	server := mcp.NewServer(
		&mcp.Implementation{
			Name:       "kfp-operator-mcp",
			Title:      "kfp operator mcp",
			Version:    "0.0.1",
			WebsiteURL: "na",
			Icons:      nil,
		},
		&mcp.ServerOptions{},
	)

	// Register resources with the server
	for _, resource := range s.resourceDefinitions() {
		server.AddResource(&resource.r, resource.h)
	}

	// Create HTTP handler for the MCP server
	handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
		return server
	}, &mcp.StreamableHTTPOptions{})

	mux := http.NewServeMux()
	mux.Handle("/mcp", handler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return http.ListenAndServe(":8000", mux)
}

//
// Resource definitions (CORRECT TYPE)
//

type ResourceHandle struct {
	r mcp.Resource
	h mcp.ResourceHandler
}

func (s *MCPServer) resourceDefinitions() []ResourceHandle {
	return []ResourceHandle{
		{
			r: mcp.Resource{
				URI:         "kfp://pipelines",
				Name:        "pipelines",
				Description: "Kubeflow Pipelines managed by the KFP Operator",
			},

			h: func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				list := &v1beta1.PipelineList{}
				if err := s.Cache.List(ctx, list); err != nil {
					return nil, err
				}

				b, err := json.Marshal(list)
				if err != nil {
					return nil, err
				}
				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{
							URI:      "kfp://pipelines",
							MIMEType: "application/json",
							Blob:     b,
						},
					},
				}, nil
			},
		},
	}
}
