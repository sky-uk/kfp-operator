package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	"github.com/sky-uk/kfp-operator/apis/pipelines"
	"github.com/sky-uk/kfp-operator/argo/common"
	"io"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	HttpHeaderContentType = "Content-Type"
	HttpContentTypeJSON   = "application/json"
)

type RunCompletionFeed struct {
	ctx            context.Context
	eventProcessor EventProcessor
	upstreams      []UpstreamService
}

func NewRunCompletionFeed(ctx context.Context, client client.Reader, endpoints []config.Endpoint) RunCompletionFeed {
	eventProcessor := NewResourceArtifactsEventProcessor(client)

	upstreams := pipelines.Map(endpoints, func(endpoint config.Endpoint) UpstreamService {
		return NewGrpcTrigger(ctx, endpoint)
	})

	return RunCompletionFeed{
		ctx:            ctx,
		eventProcessor: eventProcessor,
		upstreams:      upstreams,
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

type DataWrapper struct {
	Data common.RunCompletionEventData `json:"data"`
}

func (rcf RunCompletionFeed) extractRunCompletionEvent(request *http.Request) (*common.RunCompletionEvent, error) {
	body, err := getRequestBody(rcf.ctx, request)
	if err != nil {
		return nil, err
	}

	runDataWrapper := &DataWrapper{}
	if err := json.Unmarshal(body, runDataWrapper); err != nil {
		return nil, err
	}

	rce, err := rcf.eventProcessor.ToRunCompletionEvent(rcf.ctx, runDataWrapper.Data)
	if err != nil {
		return nil, err
	} else if rce == nil {
		return nil, errors.New("event data is empty")
	} else {
		return rce, nil
	}
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
		event, err := rcf.extractRunCompletionEvent(request)
		if err != nil || event == nil {
			logger.Error(err, "Failed to extract body from request")
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		} else {
			for _, upstream := range rcf.upstreams {
				err := upstream.call(rcf.ctx, *event)
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
