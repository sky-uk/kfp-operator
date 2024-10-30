package pkg

import "github.com/sky-uk/kfp-operator/argo/common"

type OnCompleteHandlers struct {
	OnSuccessHandler func()
	OnFailureHandler func()
}

func (onc OnCompleteHandlers) OnSuccess() {
	if onc.OnSuccessHandler != nil {
		onc.OnSuccessHandler()
	}
}

func (onc OnCompleteHandlers) OnFailure() {
	if onc.OnFailureHandler != nil {
		onc.OnFailureHandler()
	}
}

type StreamMessage struct {
	RunCompletionEventData common.RunCompletionEventData
	OnCompleteHandlers
}
