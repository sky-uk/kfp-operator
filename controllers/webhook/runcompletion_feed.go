package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"io"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type EventData struct {
	// Header is the http request header
	Header http.Header `json:"header"`
	// Body is http request body
	Body json.RawMessage `json:"body"`
}

type RunCompletionFeed struct {
	ctx       context.Context
	upstreams []UpstreamService
}

func NewRunCompletionFeed(ctx context.Context, endpoints []config.Endpoint) RunCompletionFeed {
	upstreams := pipelines.Map(endpoints, func(endpoint config.Endpoint) UpstreamService {
		return HttpWebhook{
			Upstream: endpoint,
		}
	})
	return RunCompletionFeed{
		ctx:       ctx,
		upstreams: upstreams,
	}
}

func getRequestBody(request *http.Request) ([]byte, error) {
	if request.Body == nil {
		return nil, errors.New("request body is nil")
	}
	body, err := io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(body))
	if err != nil || len(body) == 0 {
		if err == nil {
			err = errors.New("request body is empty")
		}
		return nil, fmt.Errorf("failed to read request body, %w", err)
	}
	return body, nil
}

func extractEventData(request *http.Request) (*EventData, error) {
	body, err := getRequestBody(request)
	if err != nil {
		return nil, err
	}
	rawMessage := json.RawMessage(body)
	return &EventData{
		Header: request.Header,
		Body:   rawMessage,
	}, nil
}

func (rcf RunCompletionFeed) handleEvent(response http.ResponseWriter, request *http.Request) {
	logger := log.FromContext(rcf.ctx)
	switch request.Method {
	case http.MethodPost:
		if request.Header.Get("Content-Type") != "application/json" {
			http.Error(response, "invalid Content-Type, want `application/json`", http.StatusUnsupportedMediaType)
			return
		}
		eventData, err := extractEventData(request)
		if err != nil || eventData == nil {
			logger.Error(err, "Failed to extract run completion event")
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		} else {
			for _, upstream := range rcf.upstreams {
				err := upstream.call(rcf.ctx, *eventData)
				if err != nil {
					// upstream call failed
					http.Error(response, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			return
		}
	default:
		http.Error(response, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (rcf RunCompletionFeed) start() error {
	http.HandleFunc("/events", rcf.handleEvent)

	return http.ListenAndServe(":8080", nil)
}
