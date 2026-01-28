package mcp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	v1beta1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"go.uber.org/zap"
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

// MCPServer represents the MCP server
type MCPServer struct {
	Cache    cache.Cache
	upgrader websocket.Upgrader
}

// NewMCPServer constructs a server
func NewMCPServer(c cache.Cache) *MCPServer {
	return &MCPServer{
		Cache: c,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true }, // adjust for production
		},
	}
}

// OpenAI MCP JSON structures
type JSONRPCResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *RPCError `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// MCP-specific structures for OpenAI spec
type MCPConfig struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    Capabilities `json:"capabilities"`
	ServerInfo      ServerInfo   `json:"serverInfo"`
	Streamable      bool         `json:"streamable"`
}

type Capabilities struct {
	Resources map[string]any `json:"resources"`
	Tools     map[string]any `json:"tools"`
}

type ServerInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type MCPResource struct {
	Kind       string `json:"kind"`
	Plural     string `json:"plural"`
	Group      string `json:"group"`
	Version    string `json:"version"`
	Namespaced bool   `json:"namespaced"`
}

// Start runs the MCP server
func (s *MCPServer) Start() error {
	mux := http.NewServeMux()

	// HTTP POST endpoint for OpenWebUI verification
	mux.HandleFunc("/mcp", s.mcpVerifyHandler)

	// WebSocket endpoint for streaming
	mux.HandleFunc("/mcp/ws", s.mcpWebSocketHandler)

	// Optional health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	return http.ListenAndServe(":8000", mux)
}

// POST /mcp for OpenWebUI verification
func (s *MCPServer) mcpVerifyHandler(w http.ResponseWriter, r *http.Request) {

	rawLogger, _ := zap.NewProduction()
	log := zapr.NewLogger(rawLogger).WithName("mcpVerifyHandler")

	log.Info("Received MCP verification request")

	// Decode the incoming JSON-RPC request to get the ID
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error(err, "Failed to decode MCP verification request")
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	log.Info("received MCP verification request", "req", req)

	if r.Method != http.MethodPost {
		log.Info("Invalid HTTP method for MCP verification")
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID, // use the ID from the request
		Result: MCPConfig{
			ProtocolVersion: "0.1",
			Streamable:      true,
			Capabilities: Capabilities{
				Resources: map[string]any{
					"pipelines": map[string]any{
						"kind": "Pipeline",
					},
				},
				Tools: map[string]any{},
			},
			ServerInfo: ServerInfo{
				Name:        "KFP-Operator MCP Server",
				Description: "MCP server exposing Kubeflow Pipelines",
				Version:     "0.0.1",
			},
		},
	}

	json.NewEncoder(w).Encode(resp)
}

// WebSocket handler for streaming JSON-RPC
func (s *MCPServer) mcpWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

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

// JSON-RPC request struct
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Send MCP resources over WebSocket
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
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  res,
	}
	conn.WriteJSON(resp)
}

// Send pipelines from cache over WebSocket
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

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  pipelineList.Items,
	}
	conn.WriteJSON(resp)
}
