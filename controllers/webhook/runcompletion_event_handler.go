package webhook

import (
	"github.com/sky-uk/kfp-operator/argo/common"
)

type RunCompletionEventHandler interface {
	Handle(event common.RunCompletionEvent) EventError
}
