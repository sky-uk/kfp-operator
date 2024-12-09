package webhook

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type RunCompletionEventHandler interface {
	Handle(ctx context.Context, event common.RunCompletionEvent) error
}
