package resource

import "context"

type HttpHandledResource interface {
	Type() string
	Create(ctx context.Context, body []byte) (ResponseBody, error)
	Update(ctx context.Context, id string, body []byte) (ResponseBody, error)
	Delete(ctx context.Context, id string) error
}

type ResponseBody struct {
	Id            string `json:"id,omitempty"`
	ProviderError string `json:"providerError,omitempty"`
}
