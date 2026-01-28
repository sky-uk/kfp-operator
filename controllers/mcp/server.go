package mcp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-logr/logr"
	v1beta1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/cache"
)

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

type MCPServer struct {
	Cache cache.Cache
}

type MCPConfig struct {
	Name         string          `json:"name"`
	Version      string          `json:"version"`
	Capabilities map[string]bool `json:"capabilities"`
}

type MCPResource struct {
	Kind       string `json:"kind"`
	Plural     string `json:"plural"`
	Group      string `json:"group"`
	Version    string `json:"version"`
	Namespaced bool   `json:"namespaced"`
}

func (s *MCPServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/mcp", s.mcpHandler)
	mux.HandleFunc("/mcp/resources", s.mcpResourcesHandler)
	mux.HandleFunc("/mcp/resources/pipelines", s.listPipelines)

	return http.ListenAndServe(":8000", mux)
}

func (s *MCPServer) mcpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	marshal, err := json.Marshal(MCPConfig{
		Name:    "KFP-Operator MCP Server",
		Version: "0.0.1",
		Capabilities: map[string]bool{
			"resources": true,
			"tools":     true,
			"streaming": false,
		},
	})
	if err != nil {
		return
	}
	_, err = w.Write(marshal)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (s *MCPServer) mcpResourcesHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	marshalled, err := json.Marshal([]MCPResource{
		{
			Kind:       "Pipeline",
			Plural:     "Pipelines",
			Group:      "pipelines.kubeflow.org",
			Version:    "v1beta1",
			Namespaced: true,
		},
	})
	if err != nil {
		http.Error(w, "Failed to marshal resources", http.StatusInternalServerError)
		return
	}
	_, err = w.Write(marshalled)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func (s *MCPServer) listPipelines(w http.ResponseWriter, r *http.Request) {
	pipelineList := &v1beta1.PipelineList{}
	err := s.Cache.List(r.Context(), pipelineList)
	if err != nil {
		http.Error(w, "Failed to list pipelines", http.StatusInternalServerError)
		return
	}

	body, err := json.Marshal(pipelineList.Items)
	if err != nil {
		http.Error(w, "Failed to marshal pipelines", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(body)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
