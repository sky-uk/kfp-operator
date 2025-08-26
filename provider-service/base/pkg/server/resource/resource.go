package resource

import (
	"context"
	"github.com/sky-uk/kfp-operator/pkg/providers/base"
)

type HttpHandledResource interface {
	Type() string
	Create(ctx context.Context, body []byte) (base.Output, error)
	Update(ctx context.Context, id string, body []byte) (base.Output, error)
	Delete(ctx context.Context, id string) error
}
