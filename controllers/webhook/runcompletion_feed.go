package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	pipelineshub "github.com/sky-uk/kfp-operator/apis/pipelines/hub"
	"github.com/sky-uk/kfp-operator/argo/common"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	HttpHeaderContentType = "Content-Type"
	HttpContentTypeJSON   = "application/json"
)

type RunCompletionFeed struct {
	client         client.Reader
	eventProcessor EventProcessor
	eventHandlers  []RunCompletionEventHandler
}

func NewRunCompletionFeed(
	client client.Reader,
	handlers []RunCompletionEventHandler,
) RunCompletionFeed {
	eventProcessor := NewResourceArtifactsEventProcessor()

	return RunCompletionFeed{
		client:         client,
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

func (rcf RunCompletionFeed) extractRunCompletionEventData(ctx context.Context, request *http.Request) (*common.RunCompletionEventData, EventError) {
	body, err := getRequestBody(ctx, request)
	if err != nil {
		return nil, &InvalidEvent{err.Error()}
	}

	rced := &common.RunCompletionEventData{}
	if err := json.Unmarshal(body, &rced); err != nil {
		return nil, &FatalError{err.Error()}
	}

	return rced, nil
}

func (rcf RunCompletionFeed) fetchRunConfiguration(ctx context.Context, name *common.NamespacedName) (*pipelineshub.RunConfiguration, EventError) {
	logger := common.LoggerFromContext(ctx)
	if name != nil {
		rc := &pipelineshub.RunConfiguration{}
		if err := rcf.client.Get(ctx, client.ObjectKey{
			Namespace: name.Namespace,
			Name:      name.Name,
		}, rc); err != nil {
			logger.Error(err, "failed to load", "RunConfig", name)
			if k8sErrors.IsNotFound(err) {
				return nil, &MissingResourceError{err.Error()}
			}
			return nil, &FatalError{err.Error()}
		}
		return rc, nil
	}

	return nil, nil
}

func (rcf RunCompletionFeed) fetchRun(ctx context.Context, name *common.NamespacedName) (*pipelineshub.Run, EventError) {
	logger := common.LoggerFromContext(ctx)
	if name != nil {
		run := &pipelineshub.Run{}
		if err := rcf.client.Get(ctx, client.ObjectKey{
			Namespace: name.Namespace,
			Name:      name.Name,
		}, run); err != nil {
			logger.Error(err, "failed to load", "Run", name)
			if k8sErrors.IsNotFound(err) {
				return nil, &MissingResourceError{err.Error()}
			}
			return nil, &FatalError{err.Error()}
		}
		return run, nil
	}

	return nil, nil
}

func (rcf RunCompletionFeed) HandleEvent(ctx context.Context) func(responseWriter http.ResponseWriter, request *http.Request) {
	logger := log.FromContext(ctx)

	return func(responseWriter http.ResponseWriter, request *http.Request) {
		switch request.Method {
		case http.MethodPost:

			if request.Header.Get(HttpHeaderContentType) != HttpContentTypeJSON {
				logger.Error(errors.New("RunCompletionFeed call failed"), fmt.Sprintf("invalid %s [%s], want `%s`", HttpHeaderContentType, request.Header.Get(HttpHeaderContentType), HttpContentTypeJSON))
				http.Error(responseWriter, fmt.Sprintf("invalid %s, want `%s`", HttpHeaderContentType, HttpContentTypeJSON), http.StatusUnsupportedMediaType)
				return
			}

			eventData, err := rcf.extractRunCompletionEventData(ctx, request)
			if err != nil {
				err.SendHttpError(responseWriter)
				return
			}

			var runConfiguration *pipelineshub.RunConfiguration
			var runConfigurationErr EventError
			if eventData.RunConfigurationName != nil && eventData.RunConfigurationName.Name != "" {
				runConfiguration, runConfigurationErr = rcf.fetchRunConfiguration(ctx, eventData.RunConfigurationName)
			}

			var run *pipelineshub.Run
			var runErr EventError
			if eventData.RunName != nil && eventData.RunName.Name != "" {
				run, runErr = rcf.fetchRun(ctx, eventData.RunName)
			}

			// If none are found then return error
			if run == nil && runConfiguration == nil {
				if runConfigurationErr != nil {
					runConfigurationErr.SendHttpError(responseWriter)
					return
				}
				if runErr != nil {
					runErr.SendHttpError(responseWriter)
					return
				}
			}

			event, err := rcf.eventProcessor.ToRunCompletionEvent(eventData, runConfiguration, run)
			if err != nil {
				err.SendHttpError(responseWriter)
				return
			}

			for _, handler := range rcf.eventHandlers {
				err := handler.Handle(ctx, *event)
				if err != nil {
					logger.Error(err, "Run completion event handler operation failed")
					err.SendHttpError(responseWriter)
					return
				}
			}
			return
		default:
			logger.Error(errors.New("RunCompletionFeed call failed"), "Invalid http method used [%s], only POST supported", request.Method)
			http.Error(responseWriter, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

type ObservedRunCompletionFeed struct {
	delegate        RunCompletionFeed
	requestsCounter metric.Int64Counter
}

func NewObservedRunCompletionFeed(
	client client.Reader,
	handlers []RunCompletionEventHandler,
) (ObservedRunCompletionFeed, error) {
	delegateRunCompletionFeed := NewRunCompletionFeed(client, handlers)

	meter := otel.Meter("run_completion_feed")
	requestsCounter, err := meter.Int64Counter(
		"run_completion_feed_requests",
		metric.WithDescription("Total number of requests received by the run completion feed"),
	)

	if err != nil {
		return ObservedRunCompletionFeed{}, fmt.Errorf("failed to create requests counter: %w", err)
	}

	return ObservedRunCompletionFeed{
		delegate:        delegateRunCompletionFeed,
		requestsCounter: requestsCounter,
	}, nil
}

func (orcf ObservedRunCompletionFeed) HandleEvent(ctx context.Context) func(responseWriter http.ResponseWriter, request *http.Request) {
	delegateHandler := orcf.delegate.HandleEvent(ctx)

	return func(responseWriter http.ResponseWriter, request *http.Request) {
		orcf.requestsCounter.Add(ctx, 1, metric.WithAttributes(
			attribute.String("method", request.Method),
			attribute.String("endpoint", "/events"),
		))

		delegateHandler(responseWriter, request)
	}
}
