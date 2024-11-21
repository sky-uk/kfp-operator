package webhook

import (
	"context"
	"github.com/sky-uk/kfp-operator/argo/common"
)

type RunCompletionEventHandler interface {
	handle(ctx context.Context, event common.RunCompletionEvent) error
}
