package mcpa

import (
	"context"
	"encoding/json"
	"net/http"

	v1beta1 "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/go-logr/logr"

	mcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

func (s *MCPServer) ListProviders(ctx context.Context) ([]byte, error) {
	list := &v1beta1.ProviderList{}
	if err := s.Cache.List(ctx, list); err != nil {
		return nil, err
	}

	b, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *MCPServer) ListExperiments(ctx context.Context) ([]byte, error) {
	list := &v1beta1.ExperimentList{}
	if err := s.Cache.List(ctx, list); err != nil {
		return nil, err
	}

	b, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *MCPServer) ListRunSchedules(ctx context.Context) ([]byte, error) {
	list := &v1beta1.RunScheduleList{}
	if err := s.Cache.List(ctx, list); err != nil {
		return nil, err
	}

	b, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *MCPServer) ListRunConfigurations(ctx context.Context) ([]byte, error) {
	list := &v1beta1.RunConfigurationList{}
	if err := s.Cache.List(ctx, list); err != nil {
		return nil, err
	}

	b, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *MCPServer) ListRuns(ctx context.Context) ([]byte, error) {
	list := &v1beta1.RunList{}
	if err := s.Cache.List(ctx, list); err != nil {
		return nil, err
	}

	b, err := json.Marshal(list)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *MCPServer) getOperatorLogs(
	ctx context.Context,
	namespace string,
	labelSelector string,
	container string,
	tailLines int64,
) (string, error) {

	cfg, err := rest.InClusterConfig()
	if err != nil {
		return "", err
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return "", err
	}

	pods, err := clientset.CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{
			LabelSelector: labelSelector,
		})
	if err != nil {
		return "", err
	}

	if len(pods.Items) == 0 {
		return "no matching pods found", nil
	}

	var logs string

	for _, pod := range pods.Items {
		opts := &corev1.PodLogOptions{
			Container: container,
			TailLines: &tailLines,
		}

		req := clientset.CoreV1().
			Pods(namespace).
			GetLogs(pod.Name, opts)

		stream, err := req.Stream(ctx)
		if err != nil {
			continue
		}

		buf, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			continue
		}

		logs += "=== pod: " + pod.Name + " ===\n"
		logs += string(buf) + "\n"
	}

	if logs == "" {
		return "logs unavailable", nil
	}

	return logs, nil
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
					IsError: false,
				}, nil
			},
		},
		{
			t: mcp.Tool{
				Description: "List Providers managed by the KFP Operator",
				Name:        "list_providers",
				Title:       "List Providers",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
			},
			h: func(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				providersJson, err := s.ListProviders(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.CallToolResult{
					Meta: nil,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text:        string(providersJson),
							Meta:        nil,
							Annotations: nil,
						},
					},
					IsError: false,
				}, nil
			},
		},
		{
			t: mcp.Tool{
				Description: "List Experiments managed by the KFP Operator",
				Name:        "list_experiments",
				Title:       "List Experiments",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
			},
			h: func(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				experimentsJson, err := s.ListExperiments(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.CallToolResult{
					Meta: nil,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text:        string(experimentsJson),
							Meta:        nil,
							Annotations: nil,
						},
					},
					IsError: false,
				}, nil
			},
		},
		{
			t: mcp.Tool{
				Description: "List RunSchedules managed by the KFP Operator",
				Name:        "list_runschedules",
				Title:       "List RunSchedules",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
			},
			h: func(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				runSchedulesJson, err := s.ListRunSchedules(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.CallToolResult{
					Meta: nil,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text:        string(runSchedulesJson),
							Meta:        nil,
							Annotations: nil,
						},
					},
					IsError: false,
				}, nil
			},
		},
		{
			t: mcp.Tool{
				Description: "List RunConfigurations managed by the KFP Operator",
				Name:        "list_runconfigurations",
				Title:       "List RunConfigurations",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
			},
			h: func(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				runConfigurationsJson, err := s.ListRunConfigurations(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.CallToolResult{
					Meta: nil,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text:        string(runConfigurationsJson),
							Meta:        nil,
							Annotations: nil,
						},
					},
					IsError: false,
				}, nil
			},
		},
		{
			t: mcp.Tool{
				Description: "List Runs managed by the KFP Operator",
				Name:        "list_runs",
				Title:       "List Runs",
				InputSchema: map[string]interface{}{
					"type": "object",
				},
			},
			h: func(ctx context.Context, _ *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				runsJson, err := s.ListRuns(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.CallToolResult{
					Meta: nil,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text:        string(runsJson),
							Meta:        nil,
							Annotations: nil,
						},
					},
					IsError: false,
				}, nil
			},
		},
		{
			t: mcp.Tool{
				Name:        "get_operator_logs",
				Title:       "Get Operator Logs",
				Description: "Fetch logs from operator-managed components",
				InputSchema: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"component": map[string]interface{}{
							"type":        "string",
							"description": "Operator component name (controller, webhook, eventbus, provider, run-completion-event-trigger)",
							"enum":        []string{"controller", "webhook", "eventbus", "provider", "run-completion-event-trigger"},
						},
						"namespace": map[string]interface{}{
							"type":        "string",
							"default":     "kfp-system",
							"description": "Kubernetes namespace where the component is running",
						},
						"tailLines": map[string]interface{}{
							"type":        "integer",
							"default":     200,
							"description": "Number of log lines to return",
						},
					},
					"required": []string{"component"},
				},
			},
			h: func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
				// Parse arguments from the raw JSON
				var args struct {
					Component string `json:"component"`
					Namespace string `json:"namespace"`
					TailLines int64  `json:"tailLines"`
				}
				args.Namespace = "kfp-system" // default value
				args.TailLines = 200          // default value

				if err := json.Unmarshal(req.Params.Arguments, &args); err != nil {
					return &mcp.CallToolResult{
						IsError: true,
						Content: []mcp.Content{
							&mcp.TextContent{Text: "failed to parse arguments: " + err.Error()},
						},
					}, nil
				}

				component := args.Component
				namespace := args.Namespace
				tailLines := args.TailLines

				// 🔒 operator-owned mapping (NO arbitrary pod access)
				var labelSelector string
				var container string

				switch component {
				case "controller":
					labelSelector = "control-plane=controller-manager"
					container = "manager"
				case "webhook":
					labelSelector = "control-plane=controller-manager"
					container = "manager"
				case "eventbus":
					labelSelector = "app=kfp-operator-events"
					container = "stan"
				case "provider":
					labelSelector = "app=provider-vai"
					container = "provider-service"
				case "run-completion-event-trigger":
					labelSelector = "app=run-completion-event-trigger"
					container = "run-completion-event-trigger"
				default:
					return &mcp.CallToolResult{
						IsError: true,
						Content: []mcp.Content{
							&mcp.TextContent{
								Text: "unknown component: " + component,
							},
						},
					}, nil
				}

				logs, err := s.getOperatorLogs(
					ctx,
					namespace,
					labelSelector,
					container,
					tailLines,
				)
				if err != nil {
					return nil, err
				}

				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: logs,
						},
					},
					IsError: false,
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
		{
			r: mcp.Resource{
				URI:         "kfp://providers",
				Name:        "providers",
				Description: "Providers managed by the KFP Operator",
			},

			h: func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				providersJson, err := s.ListProviders(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{
							URI:      "kfp://providers",
							MIMEType: "application/json",
							Text:     string(providersJson),
						},
					},
				}, nil
			},
		},
		{
			r: mcp.Resource{
				URI:         "kfp://experiments",
				Name:        "experiments",
				Description: "Experiments managed by the KFP Operator",
			},

			h: func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				experimentsJson, err := s.ListExperiments(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{
							URI:      "kfp://experiments",
							MIMEType: "application/json",
							Text:     string(experimentsJson),
						},
					},
				}, nil
			},
		},
		{
			r: mcp.Resource{
				URI:         "kfp://runschedules",
				Name:        "runschedules",
				Description: "RunSchedules managed by the KFP Operator",
			},

			h: func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				runSchedulesJson, err := s.ListRunSchedules(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{
							URI:      "kfp://runschedules",
							MIMEType: "application/json",
							Text:     string(runSchedulesJson),
						},
					},
				}, nil
			},
		},
		{
			r: mcp.Resource{
				URI:         "kfp://runconfigurations",
				Name:        "runconfigurations",
				Description: "RunConfigurations managed by the KFP Operator",
			},

			h: func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				runConfigurationsJson, err := s.ListRunConfigurations(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{
							URI:      "kfp://runconfigurations",
							MIMEType: "application/json",
							Text:     string(runConfigurationsJson),
						},
					},
				}, nil
			},
		},
		{
			r: mcp.Resource{
				URI:         "kfp://runs",
				Name:        "runs",
				Description: "Runs managed by the KFP Operator",
			},

			h: func(ctx context.Context, _ *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
				runsJson, err := s.ListRuns(ctx)
				if err != nil {
					return nil, err
				}

				return &mcp.ReadResourceResult{
					Contents: []*mcp.ResourceContents{
						{
							URI:      "kfp://runs",
							MIMEType: "application/json",
							Text:     string(runsJson),
						},
					},
				}, nil
			},
		},
	}
}
