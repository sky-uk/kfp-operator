package resource

type HttpHandledResource interface {
	Type() string
	Create(body []byte) (ResponseBody, error)
	Update(id string, body []byte) error
	Delete(id string) error
}
