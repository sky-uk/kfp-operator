package mcp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	v1beta1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/gorilla/websocket"
)

// Runnable starts the MCP server in a goroutine
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

// MCPServer represents the WebSocket MCP server
type MCPServer struct {
	Cache    cache.Cache
	upgrader websocket.Upgrader
}

// NewMCPServer constructs a server
func NewMCPServer(c cache.Cache) *MCPServer {
	return &MCPServer{
		Cache: c,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // adjust in prod
		},
	}
}

// JSON-RPC 2.0 structures
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      string          `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      string    `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OpenAI MCP structures
type MCPConfig struct {
	Name            string       `json:"name"`
	Version         string       `json:"version"`
	Streamable      bool         `json:"streamable"`
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
}

type Capabilities struct {
	Resources map[string]any `json:"resources"`
	Tools     map[string]any `json:"tools"`
}

type ServerInfo struct {
	Description string `json:"description"`
}

type MCPResource struct {
	Kind       string `json:"kind"`
	Plural     string `json:"plural"`
	Group      string `json:"group"`
	Version    string `json:"version"`
	Namespaced bool   `json:"namespaced"`
}

// Start runs the WebSocket MCP server
func (s *MCPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.mcpWebSocketHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return http.ListenAndServe(":8000", mux)
}

// WebSocket handler for MCP
func (s *MCPServer) mcpWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Initial handshake with OpenAI MCP spec
	initial := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      "VERIFY_TOOL_SERVER",
		Result: MCPConfig{
			Name:            "KFP-Operator MCP Server",
			Version:         "0.0.1",
			Streamable:      true,
			ProtocolVersion: "0.1",
			Capabilities: Capabilities{
				Resources: map[string]any{
					"pipelines": map[string]any{"kind": "Pipeline"},
				},
				Tools: map[string]any{},
			},
			ServerInfo: ServerInfo{
				Description: "KFP Operator MCP server compliant with OpenAI spec",
			},
		},
	}

	if err := conn.WriteJSON(initial); err != nil {
		return
	}

	// Handle incoming JSON-RPC requests
	for {
		var req JSONRPCRequest
		if err := conn.ReadJSON(&req); err != nil {
			break // client disconnected
		}

		switch req.Method {
		case "LIST_RESOURCES":
			s.sendResources(conn, req.ID)
		case "LIST_PIPELINES":
			s.sendPipelines(conn, req.ID, r.Context())
		default:
			conn.WriteJSON(JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Error: &RPCError{
					Code:    -32601,
					Message: "Method not found",
				},
			})
		}
	}
}

// Send MCP resources
func (s *MCPServer) sendResources(conn *websocket.Conn, id string) {
	res := []MCPResource{
		{
			Kind:       "Pipeline",
			Plural:     "Pipelines",
			Group:      "pipelines.kubeflow.org",
			Version:    "v1beta1",
			Namespaced: true,
		},
	}
	conn.WriteJSON(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  res,
	})
}

// Send pipelines from cache
func (s *MCPServer) sendPipelines(conn *websocket.Conn, id string, ctx context.Context) {
	pipelineList := &v1beta1.PipelineList{}
	if err := s.Cache.List(ctx, pipelineList); err != nil {
		conn.WriteJSON(JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      id,
			Error: &RPCError{
				Code:    -32000,
				Message: "Failed to list pipelines",
			},
		})
		return
	}

	conn.WriteJSON(JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  pipelineList.Items,
	})
}
