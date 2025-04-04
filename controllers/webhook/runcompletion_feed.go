package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/sky-uk/kfp-operator/argo/common"
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
	eventHandlers  []RunCompletionEventHandler
}

func NewRunCompletionFeed(
	ctx context.Context,
	client client.Reader,
	handlers []RunCompletionEventHandler,
) RunCompletionFeed {
	eventProcessor := NewResourceArtifactsEventProcessor(client)
	return RunCompletionFeed{
		ctx:            ctx,
		eventProcessor: eventProcessor,
		eventHandlers:  handlers,
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

func (rcf RunCompletionFeed) extractRunCompletionEvent(request *http.Request) (*common.RunCompletionEvent, error) {
	body, err := getRequestBody(rcf.ctx, request)
	if err != nil {
		return nil, err
	}

	rced := common.RunCompletionEventData{}
	if err := json.Unmarshal(body, &rced); err != nil {
		return nil, err
	}

	rce, err := rcf.eventProcessor.ToRunCompletionEvent(rcf.ctx, rced)
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
			logger.Error(errors.New("RunCompletionFeed call failed"), fmt.Sprintf("invalid %s [%s], want `%s`", HttpHeaderContentType, request.Header.Get(HttpHeaderContentType), HttpContentTypeJSON))
			http.Error(response, fmt.Sprintf("invalid %s, want `%s`", HttpHeaderContentType, HttpContentTypeJSON), http.StatusUnsupportedMediaType)
			return
		}
		event, err := rcf.extractRunCompletionEvent(request)
		if err != nil || event == nil {
			logger.Error(err, "Failed to extract body from request")
			http.Error(response, err.Error(), http.StatusBadRequest)
			return
		} else {
			for _, handler := range rcf.eventHandlers {
				err := handler.Handle(*event)
				if err != nil {
					logger.Error(err, "Run completion event handler operation failed")
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
