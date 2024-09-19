package webhook

import (
	"bytes"
	"context"
	config "github.com/sky-uk/kfp-operator/apis/config/v1alpha5"
	"net"
	"net/http"
	"time"
)

type UpstreamService interface {
	call(ctx context.Context, ed EventData) error
}

type HttpWebhook struct {
	Endpoint config.Endpoint
	Client   *http.Client
}

func (hw HttpWebhook) buildRequest(ctx context.Context, bodyBytes []byte) (*http.Request, error) {
	return http.NewRequestWithContext(ctx, http.MethodPost, hw.Endpoint.Path, bytes.NewReader(bodyBytes))
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
		Endpoint: endpoint,
		Client:   client,
	}
}

func (hw HttpWebhook) call(ctx context.Context, ed EventData) error {
	bodyBytes, err := ed.Body.MarshalJSON()
	if err != nil {
		return err
	}
	req, err := hw.buildRequest(ctx, bodyBytes)
	if err != nil {
		return err
	}
	_, err = hw.Client.Do(req)
	if err != nil {
		return err
	}
	// switch on response code and error on anything other than 200
	return nil
}
