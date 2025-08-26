package webhook

import (
	"context"
	"github.com/sky-uk/kfp-operator/pkg/common"
)

type RunCompletionEventHandler interface {
	Handle(ctx context.Context, event common.RunCompletionEvent) EventError
}
