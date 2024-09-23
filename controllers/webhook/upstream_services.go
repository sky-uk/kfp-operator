package webhook

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	"github.com/sky-uk/kfp-operator/argo/common"
	"io"
	"net"
	"net/http"
	"time"
)

type UpstreamService interface {
	call(ctx context.Context, ed EventData) error
}

type HttpWebhook struct {
	Upstream config.Endpoint
	Client   *http.Client
}

func NewHttpWebhook(endpoint config.Endpoint) HttpWebhook {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: 2 * time.Second, // Timeout for establishing connection
		}).DialContext,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second, // Total timeout for the request
	}

	return HttpWebhook{
		Upstream: endpoint,
		Client:   client,
	}
}

func (hw HttpWebhook) buildRequest(ctx context.Context, bodyBytes []byte) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodPost, hw.Upstream.URL(), bytes.NewReader(bodyBytes))
}

func (hw HttpWebhook) transferHeaders(headers http.Header, req *http.Request) {
	for headerName, headerValues := range headers {
		for _, headerValue := range headerValues {
			req.Header.Add(headerName, headerValue)
		}
	}
}

func (hw HttpWebhook) call(ctx context.Context, ed EventData) error {
	logger := common.LoggerFromContext(ctx)
	bodyBytes, err := ed.Body.MarshalJSON()
	if err != nil {
		return err
	}

	req, err := hw.buildRequest(ctx, bodyBytes)
	if err != nil {
		return err
	}
	hw.transferHeaders(ed.Header, req)

	response, err := hw.Client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error(err, "Failed to close body")
		}
	}(response.Body)

	// Fully consume the response body
	_, err = io.Copy(io.Discard, response.Body)
	if err != nil {
		logger.Error(err, "Failed to fully consume response body")
	}

	if response.StatusCode/100 != 2 {
		return errors.New(fmt.Sprintf("Upstream service returned error, http status code: [%s]", response.Status))
	}
	return nil
}
