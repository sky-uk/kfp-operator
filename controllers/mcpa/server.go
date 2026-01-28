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

	for _, tool := range s.tools() {
		server.AddTool(&tool.t, tool.h)
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
	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./controllers/mcpa/static/openapi.json")
	})
	mux.HandleFunc("/openai.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./controllers/mcpa/static/openai.json")
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

type ToolHandle struct {
	t mcp.Tool
	h mcp.ToolHandler
}

func (s *MCPServer) ListPipelines(ctx context.Context) ([]byte, error) {
	list := &v1beta1.PipelineList{}
	if err := s.Cache.List(ctx, list); err != nil {
		return nil, err
	}

	b, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *MCPServer) tools() []ToolHandle {
	return []ToolHandle{
		{
			t: mcp.Tool{
				Description: "List Kubeflow Pipelines managed by the KFP Operator",
				Name:        "list_pipelines",
				Title:       "List Pipelines",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
				OutputSchema: map[string]interface{}{
					"type": "object",
				},
			},
			h: func(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				pipelinesJson, err := s.ListPipelines(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.CallToolResult{
					Meta: nil,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text:        string(pipelinesJson),
							Meta:        nil,
							Annotations: nil,
						},
					},
					StructuredContent: true,
					IsError:           false,
				}, nil
			},
		},
	}
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
				pipelinesJson, err := s.ListPipelines(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{
							URI:      "kfp://pipelines",
							MIMEType: "application/json",
							Text:     string(pipelinesJson),
						},
					},
				}, nil
			},
		},
	}
}
