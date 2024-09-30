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

const (
	HttpHeaderContentType = "Content-Type"
	HttpContentTypeJSON   = "application/json"
)

type EventData struct {
	Header http.Header     `json:"header"`
	Body   json.RawMessage `json:"body"`
}

type RunCompletionFeed struct {
	ctx       context.Context
	upstreams []UpstreamService
}

func NewRunCompletionFeed(ctx context.Context, endpoints []config.Endpoint) RunCompletionFeed {
	upstreams := pipelines.Map(endpoints, func(endpoint config.Endpoint) UpstreamService {
		return NewHttpWebhook(endpoint)
	})
	return RunCompletionFeed{
		ctx:       ctx,
		upstreams: upstreams,
	}
}

func getRequestBody(ctx context.Context, request *http.Request) ([]byte, error) {
	logger := log.FromContext(ctx)

	if request.Body == nil {
		return nil, errors.New("request body is nil")
	}

	// Ensure that the response body is closed after reading
	defer func(body io.ReadCloser) {
		err := body.Close()
		if err != nil {
			logger.Error(err, "Failed to close body")
		}
	}(request.Body)

	body, err := io.ReadAll(request.Body)
	request.Body = io.NopCloser(bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to read request body, %w", err)
	} else if len(body) == 0 {
		return nil, errors.New("request body is empty")
	}
	return body, nil
}

func extractEventData(ctx context.Context, request *http.Request) (*EventData, error) {
	body, err := getRequestBody(ctx, request)
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
		if request.Header.Get(HttpHeaderContentType) != HttpContentTypeJSON {
			logger.Error(errors.New("RunCompletionFeed call failed"), "invalid %s [%s], want `%s`", HttpHeaderContentType, request.Header.Get(HttpHeaderContentType), HttpContentTypeJSON)
			http.Error(response, fmt.Sprintf("invalid %s, want `%s`", HttpHeaderContentType, HttpContentTypeJSON), http.StatusUnsupportedMediaType)
			return
		}
		eventData, err := extractEventData(rcf.ctx, request)
		if err != nil || eventData == nil {
			logger.Error(err, "Failed to extract body from request")
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		} else {
			for _, upstream := range rcf.upstreams {
				err := upstream.call(rcf.ctx, *eventData)
				if err != nil {
					logger.Error(err, "Call to upstream failed")
					http.Error(response, err.Error(), http.StatusInternalServerError)
					return
				}
			}
			return
		}
	default:
		logger.Error(errors.New("RunCompletionFeed call failed"), "Invalid http method used [%s], only POST supported", request.Method)
		http.Error(response, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (rcf RunCompletionFeed) Start(port int) error {
	http.HandleFunc("/events", rcf.handleEvent)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
