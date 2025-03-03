package resource

type HttpHandledResource interface {
	Type() string
	Create(body []byte) (ResponseBody, error)
	Update(id string, body []byte) (ResponseBody, error)
	Delete(id string) error
}

type ResponseBody struct {
	Id            string `json:"id,omitempty"`
	ProviderError string `json:"providerError,omitempty"`
}
